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
	rewriteParts    string
	fourOhFourPath  string

	extensionsToCompress []string
	rewritePaths         *rewrites
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

	err := readFlags()
	if err != nil {
		fmt.Println("Error parsing flags:")
		fmt.Println(err)
		return 1
	}

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

func readFlags() error {
	const (
		defaultAssetsPath = "./public"
		defaultBinaryPath = "assets-server"
		defaultPort       = 8000
		defaultDebug      = false
		defaultLogging    = false
		defaultCompress   = ".html,.htm,.css,.js,.svg,.json,.txt,.xml,.yml,.yaml,.kml,.csv,.tsv,.webmanifest,.vtt,.vcard,.vcf,.ttc,.ttf,.rdf,.otf,.appcache,.md,.mdown,.m3u,.m3u8"
		defaultRewrites   = ""
		default404path    = ""
		assetsPathUsage   = "file path of the assets directory"
		binaryPathUsage   = "file path of the resulting binary"
		portUsage         = "TCP port from which the server will be reachable"
		debugUsage        = "enable verbose debug messages"
		loggingUsage      = "enable request logging for the server"
		compressUsage     = "comma separated list of file extensions to compress. To completely disable compression specify an empty string"
		rewritesUsage     = "comma separates list of colon separated tuples for internal rewriting requests (source:target)"
		fourOhfourUsage   = "path to custom 404 page"
	)
	// src flag
	flag.StringVar(&assetsPath, "src", defaultAssetsPath, assetsPathUsage)
	// output flag
	flag.StringVar(&binaryPath, "dest", defaultBinaryPath, binaryPathUsage)
	// port flag
	flag.IntVar(&port, "port", defaultPort, portUsage)
	// debug flag
	flag.BoolVar(&debugEnabled, "debug", defaultDebug, debugUsage)
	// logRequests flag
	flag.BoolVar(&loggingEnabled, "logging", defaultLogging, loggingUsage)
	// compress flag
	flag.StringVar(&compressFormats, "compress", defaultCompress, compressUsage)
	// rewrites flag
	flag.StringVar(&rewriteParts, "rewrites", defaultRewrites, rewritesUsage)
	// 404-path flag
	flag.StringVar(&fourOhFourPath, "404-path", default404path, fourOhfourUsage)

	flag.Parse()

	extensionsToCompress = strings.Split(compressFormats, ",")
	if rewriteParts != "" {
		var err error
		rewritePaths, err = initRewrites(rewriteParts)
		return err
	}
	return nil
}
