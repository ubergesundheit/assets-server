package main

import (
	"bufio"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
)

const tmpPackageName = "as-builder"

func debug(log interface{}) {
	if debugEnabled == true {
		fmt.Println(log)
	}
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

func createFiles() (string, error) {
	compilationDir, err := ioutil.TempDir(filepath.Join(build.Default.GOPATH, "src"), tmpPackageName)
	if err != nil {
		return "", err
	}
	debug("Compilation dir is " + compilationDir)

	f, err := os.Create(filepath.Join(compilationDir, "main.go"))
	if err != nil {
		return "", err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	_, err = fmt.Fprintf(w, mainGoTemplate, urlPath, port, path.Base(binaryPath), loggingEnabled)
	if err != nil {
		return "", err
	}
	w.Flush()

	return compilationDir, nil
}

func executeCompilation(compilationDir string) error {
	// find the go path before the statik path
	// in order to be able to use `go get` when statik is missing
	goBinPath, err := exec.LookPath("go")
	if err != nil {
		return err
	}

	statikPath, err := exec.LookPath("statik")
	if err != nil {
		return err
	}

	// compress files
	tmpAssetsPath := filepath.Join(compilationDir, path.Base(assetsPath))
	copy.Copy(assetsPath, tmpAssetsPath)
	err = compressFiles(tmpAssetsPath)
	if err != nil {
		return err
	}

	statikArgs := []string{"-src", tmpAssetsPath, "-dest", compilationDir}
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
		"-o", binaryPath,
		"-tags", "netgo", // use go network implementaton
		"-ldflags", "-s -w -extldflags -static",
		compilationDir + "/main.go",
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
