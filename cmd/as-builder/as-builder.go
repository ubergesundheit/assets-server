package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const tmpPackageName = "as-builder"

var debugEnabled = false

func debug(log interface{}) {
	if debugEnabled == true {
		fmt.Println(log)
	}
}

func readFlags() (assetsPath, binaryPath, urlPath string, port int, loggingEnabled bool) {
	const (
		defaultAssetsPath = "./public"
		defaultBinaryPath = "assets-server"
		defaultURLPath    = "/"
		defaultPort       = 8000
		assetsPathUsage   = "file path of the assets directory"
		binaryPathUsage   = "file path of the resulting binary"
		urlPathUsage      = "URL path for the server"
		portUsage         = "TCP port from which the server will be reachable"
		debugUsage        = "enable verbose debug messages"
		loggingUsage      = "enable request logging for the server"
	)
	// src flag
	flag.StringVar(&assetsPath, "src", defaultAssetsPath, assetsPathUsage)
	// output flag
	flag.StringVar(&binaryPath, "dest", defaultBinaryPath, binaryPathUsage)
	// url flag
	flag.StringVar(&urlPath, "url", defaultURLPath, urlPathUsage)
	// port flag
	flag.IntVar(&port, "port", defaultPort, portUsage)
	// debug flag
	flag.BoolVar(&debugEnabled, "debug", false, debugUsage)
	// logRequests flag
	flag.BoolVar(&loggingEnabled, "logging", false, loggingUsage)

	flag.Parse()
	return assetsPath, binaryPath, urlPath, port, loggingEnabled
}

func checkDependencies() (errs []error) {
	for _, dep := range []string{"go"} {
		_, err := exec.LookPath(dep)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func createFiles(urlPath, binaryName string, port int, loggingEnabled bool) (string, string, error) {
	compilationDir, err := ioutil.TempDir(filepath.Join(build.Default.GOPATH, "src"), tmpPackageName)
	if err != nil {
		return "", "", err
	}
	debug("Compilation dir is " + compilationDir)

	mainPath := filepath.Join(compilationDir, "main.go")
	f, err := os.Create(mainPath)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	_, err = fmt.Fprintf(w, mainGoTemplate, urlPath, port, binaryName, loggingEnabled)
	if err != nil {
		return "", "", err
	}
	w.Flush()

	return compilationDir, mainPath, nil
}

func executeCompilation(compilationDir, mainPath, assetsPath, outPath string) error {
	// find the go path before the statik path
	// in order to be able to use `go get` when statik is missing
	goBinPath, err := exec.LookPath("go")
	if err != nil {
		return err
	}

	statikPath, err := exec.LookPath("statik")
	if err != nil {
		debug("statik not found, installing it with go get")
		// if not found, install statik with `go get`
		if err := exec.Command(goBinPath, []string{"get", "-u", "github.com/rakyll/statik"}...).Run(); err != nil {
			return err
		}
		statikPath, err = exec.LookPath("statik")
		if err != nil {
			return err
		}
	}

	statikArgs := []string{"-src", assetsPath, "-dest", compilationDir}
	statikCmd := exec.Command(statikPath, statikArgs...)
	statikCmd.Stdout = os.Stdout
	statikCmd.Stderr = os.Stderr
	debug("executing " + strings.Join(statikCmd.Args, " "))
	if err := statikCmd.Run(); err != nil {
		return err
	}

	buildArgs := []string{
		"build",
		"-a", // rebuild all
		"-o", outPath,
		"-tags", "netgo", // use go network implementaton
		"-ldflags", "-s -w -extldflags -static",
		mainPath,
	}

	command := exec.Command(goBinPath, buildArgs...)
	command.Env = append(os.Environ(),
		"CGO_ENABLED=0",
	)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	debug("executing " + strings.Join(command.Args, " "))
	if err = command.Run(); err != nil {
		return err
	}
	return nil
}
