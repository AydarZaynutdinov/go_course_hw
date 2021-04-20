package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFile bool) error {
	return levelDirTree(out, path, printFile, make([]bool, 0, 0))
}

func levelDirTree(out io.Writer, path string, printFile bool, level []bool) error {
	// open current file
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	// get all his files
	allReadFiles, err := f.ReadDir(-1)
	if err != nil {
		return err
	}

	// filter not dir files if printFile is false
	allWorkFiles := filterFiles(allReadFiles, printFile)

	isLast := false
	// loop for each filtered file
	for ind, cur := range allWorkFiles[:] {
		// check if current file is last file on this layer
		if ind+1 == len(allWorkFiles) {
			isLast = true
		}

		// add this file to output
		err = PrintPath(out, cur, level, isLast)
		if err != nil {
			return err
		}

		if cur.IsDir() {
			if isLast {
				level = append(level, false)
			} else {
				level = append(level, true)
			}
			// if current file is dir run this func for this file
			err = levelDirTree(out, fmt.Sprintf("%s\\%s", path, cur.Name()), printFile, level)
			if err != nil {
				return err
			}
			level = level[:len(level) - 1]
		}
	}
	return nil
}

func filterFiles(allReadFiles []os.DirEntry, printFile bool) []os.DirEntry {
	allWorkFiles := make([]os.DirEntry, 0, cap(allReadFiles))
	for _, cur := range allReadFiles {
		if printFile || cur.IsDir() {
			allWorkFiles = append(allWorkFiles, cur)
		}
	}
	return allWorkFiles
}

func PrintPath(out io.Writer, file os.DirEntry, level []bool, isLast bool) error {
	// offset according by level
	for _, curLevel := range level {
		if curLevel {
			out.Write([]byte("│\t"))
		} else {
			out.Write([]byte("\t"))
		}
	}

	if isLast {
		out.Write([]byte("└───"))
	} else {
		out.Write([]byte("├───"))
	}
	out.Write([]byte(file.Name()))

	if !file.IsDir() {
		info, err := file.Info()
		if err != nil {
			return err
		}
		size := info.Size()
		var sizeMessage string
		if size != 0 {
			sizeMessage = fmt.Sprint(" (", size, "b)")
		} else {
			sizeMessage = fmt.Sprint(" (empty)")
		}
		out.Write([]byte(sizeMessage))
	}

	out.Write([]byte("\n"))
	return nil
}
