package main

import (
	"fmt"
	"log"
	"os"
	"path"
)

func main() {
	// parser arguments
	parseArguments()

	// find package for path
	var boxes = make(map[string]bool)

	for _, boxPath := range flags.Append.BoxPath {
		boxPath, value := pathForBoxPath(boxPath)
		boxes[boxPath] = value
		verbosef("box path %q added and exists=%t\n", boxPath, value)
	}

	// switch on the operation to perform
	switch flagsParser.Active.Name {
	case "append":
		operationAppend(boxes)
	}

	// all done
	verbosef("\n")
	verbosef("boxedRice finished successfully\n")
}

// helper function to get *build.Package for given path
func pathForBoxPath(boxPath string) (string, bool) {
	// get pwd for relative imports
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("error getting pwd (required for relative box paths): %s\n", err)
		os.Exit(1)
	}

	if path.IsAbs(boxPath) {
		boxPath = boxPath
	} else {
		boxPath = path.Join(pwd, boxPath)
	}

	result, err := exists(boxPath)
	if err != nil {
		fmt.Printf("error finding box path %q : %s\n", boxPath, err)
	}
	return boxPath, result
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func verbosef(format string, stuff ...interface{}) {
	if flags.Verbose {
		log.Printf(format, stuff...)
	}
}
