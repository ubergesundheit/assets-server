package main

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

	http.Handle("%[1]s", http.StripPrefix("%[1]s", http.FileServer(statikFS)))
	http.ListenAndServe(":%[2]d", nil)
}
`
