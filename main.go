package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func run() ([]string, error) {
	searchDir := "c:/path/to/dir"

	fileList := make([]string, 0)
	e := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		// parse files for any go gen tags

		// If gogen tag:
		// read component file and reducer file - if any tagged as do not generate, then keep them.
		// if output is the same as the current file, ignore and do not overwrite
		// otherwise, replace genned
		// delete anything not covered
		fileList = append(fileList, path)
		return err
	})

	if e != nil {
		panic(e)
	}

	for _, file := range fileList {
		fmt.Println(file)
	}

	return fileList, nil
}

func main() {
	run()
}
