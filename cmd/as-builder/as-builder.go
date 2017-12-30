package main

import (
	"bytes"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

var outPath = "./assets-server"
var assetsPath = "./public"
var port = 8000
var httpPath = "/"

var namePackage = "as-builder"

var mainGoTemplate = `package main

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

func main() {
	dir, err := ioutil.TempDir(filepath.Join(build.Default.GOPATH, "src"), namePackage)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.RemoveAll(dir)

	var qb bytes.Buffer
	fmt.Fprintf(&qb, mainGoTemplate, httpPath, httpPath, port)

	mainPath := filepath.Join(dir, "main.go")
	if err := ioutil.WriteFile(mainPath, qb.Bytes(), 0644); err != nil {
		fmt.Println(err)
		return
	}

	if err = runStatik(dir); err != nil {
		fmt.Println(err)
		return
	}
	if err = runGoCompile(dir, mainPath); err != nil {
		fmt.Println(err)
		return
	}
}

func runStatik(dir string) error {
	cmd, err := exec.LookPath("statik")
	if err != nil {
		return err
	}
	args := []string{"-src", assetsPath, "-dest", dir}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		return err
	}
	return nil
}

func runGoCompile(dir, mainPath string) error {
	cmd, err := exec.LookPath("go")
	if err != nil {
		return err
	}
	args := []string{"build", "-o", outPath, "-a", "-tags", "netgo", "-ldflags", "-extldflags -static", mainPath}

	command := exec.Command(cmd, args...)
	command.Env = append(os.Environ(),
		"CGO_ENABLED=0",
	)

	if err = command.Run(); err != nil {
		return err
	}
	return nil
}
