package main

import (
	"bytes"
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

const mainGoTemplate = `package main

import (
	"log"
	"net/http"

	_ "./statik"
	"github.com/rakyll/statik/fs"
)

func main() {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("%s", http.StripPrefix("%s", http.FileServer(statikFS)))
	http.ListenAndServe(":%d", nil)
}
`

var debugEnabled = false

func debug(log interface{}) {
	if debugEnabled == true {
		fmt.Println(log)
	}
}

func main() {
	errs := checkDependencies()
	if len(errs) != 0 {
		fmt.Println("Error: Dependencies not met:")
		for _, err := range errs {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	assetsPath, binaryPath, urlPath, port := readFlags()

	compilationDir, mainPath, err := createFiles(urlPath, port)
	if err != nil {
		fmt.Println("Error creating files:")
		fmt.Println(err)
		os.RemoveAll(compilationDir)
		os.Exit(1)
	}

	if err = executeCompilation(compilationDir, mainPath, assetsPath, binaryPath); err != nil {
		fmt.Println(err)
		os.RemoveAll(compilationDir)
		os.Exit(1)
	}
	debug("done")
}

func readFlags() (assetsPath, binaryPath, urlPath string, port int) {
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

	flag.Parse()
	return assetsPath, binaryPath, urlPath, port
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

func createFiles(urlPath string, port int) (string, string, error) {
	compilationDir, err := ioutil.TempDir(filepath.Join(build.Default.GOPATH, "src"), tmpPackageName)
	if err != nil {
		return "", "", err
	}
	debug("Compilation dir is " + compilationDir)

	var qb bytes.Buffer
	fmt.Fprintf(&qb, mainGoTemplate, urlPath, urlPath, port)

	mainPath := filepath.Join(compilationDir, "main.go")
	if err := ioutil.WriteFile(mainPath, qb.Bytes(), 0644); err != nil {
		return "", "", err
	}
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
	debug("executing " + strings.Join(statikCmd.Args, " "))
	if err := statikCmd.Run(); err != nil {
		return err
	}

	buildArgs := []string{
		"build",
		"-a", // rebuild all
		"-o", outPath,
		"-tags", "netgo", // use go network implementaton
		"-ldflags", "-extldflags -static",
		mainPath,
	}

	command := exec.Command(goBinPath, buildArgs...)
	command.Env = append(os.Environ(),
		"CGO_ENABLED=0",
	)

	debug("executing " + strings.Join(command.Args, " "))
	if err = command.Run(); err != nil {
		return err
	}
	return nil
}
