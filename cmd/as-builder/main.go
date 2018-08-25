package main

import (
	"fmt"
	"os"
	"path"
)

var exitCode = 1

func main() {
	os.Exit(execute())
}

func execute() int {
	errs := checkDependencies()
	if len(errs) != 0 {
		fmt.Println("Error: Dependencies not met:")
		for _, err := range errs {
			fmt.Println(err)
		}
		return 1
	}

	assetsPath, binaryPath, urlPath, port, loggingEnabled := readFlags()

	compilationDir, mainPath, err := createFiles(urlPath, path.Base(binaryPath), port, loggingEnabled)
	if err != nil {
		fmt.Println("Error creating files:")
		fmt.Println(err)
		return 1
	}
	defer os.RemoveAll(compilationDir)

	if err = executeCompilation(compilationDir, mainPath, assetsPath, binaryPath); err != nil {
		fmt.Println(err)
		return 1
	}
	debug("done")
	return 0
}
