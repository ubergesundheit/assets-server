package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	exitCode = 1

	debugEnabled    bool
	assetsPath      string
	binaryPath      string
	urlPath         string
	port            int
	loggingEnabled  bool
	compressFormats string

	extensionsToCompress []string
)

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

	readFlags()

	compilationDir, err := createFiles()
	if err != nil {
		fmt.Println("Error creating files:")
		fmt.Println(err)
		return 1
	}
	defer os.RemoveAll(compilationDir)

	if err = executeCompilation(compilationDir); err != nil {
		fmt.Println(err)
		return 1
	}
	debug("done")
	return 0
}

func readFlags() {
	const (
		defaultAssetsPath = "./public"
		defaultBinaryPath = "assets-server"
		defaultURLPath    = "/"
		defaultPort       = 8000
		defaultDebug      = false
		defaultLogging    = false
		defaultCompress   = ".html,.htm,.css,.js,.svg,.json,.txt,.xml,.yml,.yaml,.kml,.csv,.tsv,.webmanifest,.vtt,.vcard,.vcf,.ttc,.ttf,.rdf,.otf,.appcache,.md,.mdown,.m3u,.m3u8"
		assetsPathUsage   = "file path of the assets directory"
		binaryPathUsage   = "file path of the resulting binary"
		urlPathUsage      = "URL path for the server"
		portUsage         = "TCP port from which the server will be reachable"
		debugUsage        = "enable verbose debug messages"
		loggingUsage      = "enable request logging for the server"
		compressUsage     = "comma separated list of file extensions to compress. To completely disable compression specify an empty string"
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
	flag.BoolVar(&debugEnabled, "debug", defaultDebug, debugUsage)
	// logRequests flag
	flag.BoolVar(&loggingEnabled, "logging", defaultLogging, loggingUsage)
	// compress flag
	flag.StringVar(&compressFormats, "compress", defaultCompress, compressUsage)

	extensionsToCompress = strings.Split(compressFormats, ",")

	flag.Parse()
}
