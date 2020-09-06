package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/newm4n/go-resource/globber"
)

func main() {

	basePathPtr := flag.String("base", "", "File filter under the base path.")
	newPathPtr := flag.String("path", "", "The URL path that will replace the base path")
	filterPtr := flag.String("filter", "", "Ant based file filter under the base path. e.g /**/*")
	goPathPtr := flag.String("go", "", "Path to Go File, It will replace if exist.")
	packagePtr := flag.String("package", "", "The package name to make")

	flag.Parse()

	if basePathPtr == nil || goPathPtr == nil || packagePtr == nil || filterPtr == nil || newPathPtr == nil || len(*basePathPtr) == 0 || len(*goPathPtr) == 0 || len(*packagePtr) == 0 || len(*filterPtr) == 0 || len(*newPathPtr) == 0 {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "\t %s -base \"/path/to/base/location\" -path \"/prefix\" -filter \"/**/*\" -go \"/path/to/target/file.go\" -package fancypackage", os.Args[0])
	} else {
		fmt.Printf("Scanning base       : %s\n", *basePathPtr)
		fmt.Printf("New Base            : %s\n", *newPathPtr)
		fmt.Printf("Using filter        : %s\n", *filterPtr)
		fmt.Printf("Created go file     : %s\n", *goPathPtr)
		fmt.Printf("Created go package  : %s\n", *packagePtr)
		fmt.Println("--------------------")
		files, err := findPaths(*basePathPtr, *filterPtr)
		if err != nil || len(files) == 0 {
			fmt.Printf("ERROR : Error or no file to process :  %v\n", err)
		} else {
			fmt.Printf("Processing %d files", len(files))
			err := binIt(*goPathPtr, *packagePtr, files, *basePathPtr, *newPathPtr)
			if err != nil {
				fmt.Printf("ERROR : Error while bin it :  %v\n", err)
			}
			fmt.Println("\nDone")
		}
	}
}

func binIt(target, pack string, files []string, basepath, newpath string) error {
	var f *os.File
	if _, err := os.Stat(target); os.IsNotExist(err) {
		f, err = os.Create(target)
		if err != nil {
			return err
		}
	} else {
		err = os.Remove(target)
		if err != nil {
			return err
		}
		f, err = os.Create(target)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Writing To : %s\n", target)

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("package %s\n\n", pack))

	buffer.WriteString(fmt.Sprintf("// This file is autogenerated. Do not edit this file !!\n"))
	buffer.WriteString(fmt.Sprintf("// Generated using BINIT.\n\n"))

	mimeDetect := &MimeType{}

	mimeMap := make(map[string]string)
	for idx, src := range files {
		data, err := ioutil.ReadFile(src)
		if err != nil {
			fmt.Printf("\t%s ... Error : %v\n", src, err)
		} else {
			fmt.Printf("\n\t#%d Reading : %s ... ", idx+1, src)

			mime, err := mimeDetect.MimeForFileName(src)
			if err != nil {
				mimeMap[src] = "application/octet-stream"
			} else {
				mimeMap[src] = mime
			}
			fmt.Printf("%d bytes. mime : %s. Done", len(data), mimeMap[src])
		}
	}

	buffer.WriteString(fmt.Sprintf("// GetStaticResource returns all the stuffed static resources \n"))
	buffer.WriteString(fmt.Sprintf("func GetStaticResource() (map[string][]byte,map[string]string) {\n"))
	buffer.WriteString(fmt.Sprintf("		staticMap := make(map[string][]byte)\n"))
	buffer.WriteString(fmt.Sprintf("		staticMime := make(map[string]string)\n"))

	for src, mime := range mimeMap {
		name := fmt.Sprintf("%s%s", newpath, src[len(basepath):])
		buffer.WriteString(fmt.Sprintf("     staticMime[\"%s\"] = \"%s\"\n", name, mime))
	}

	for idx, src := range files {
		data, err := ioutil.ReadFile(src)
		if err != nil {
			fmt.Printf("\t%s ... Error : %v\n", src, err)
		} else {
			name := fmt.Sprintf("%s%s", newpath, src[len(basepath):])
			fmt.Printf("\n\t#%d Writing Entry : %s ... ", idx+1, name)

			if mimeDetect.IsAllPrintableChar(data) {
				buffer.WriteString(fmt.Sprintf("     staticMap[\"%s\"] = []byte(`", name))
				dataString := strings.Replace(string(data), "`", "` + \"`\" + `", -1)
				buffer.WriteString(dataString)
				buffer.WriteString("`)\n\n")
			} else {
				buffer.WriteString(fmt.Sprintf("     staticMap[\"%s\"] = []byte{", name))
				counter := 0
				for i, b := range data {
					if i > 0 {
						buffer.WriteString(",")
					}
					if counter > 16 {
						buffer.WriteString("\n")
						counter = 0
					}
					counter++
					//buffer.WriteString(fmt.Sprintf("%#x", b))
					buffer.WriteString(fmt.Sprintf("%d", b))
				}
				buffer.WriteString("}\n\n")
			}
			fmt.Printf("Done")
		}
	}
	buffer.WriteString("    return staticMap,staticMime\n}\n")

	write, err := f.Write(buffer.Bytes())
	if err != nil {
		return err
	}
	fmt.Printf("\nTotal written %d bytes to %s", write, target)
	return nil
}

func findPaths(dir, filter string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0)
	for _, file := range files {
		fullpath := fmt.Sprintf("%s/%s", dir, file.Name())
		if file.IsDir() {
			rfile, err := findPaths(fullpath, filter)
			if err != nil || rfile == nil {
				fmt.Printf("Returning from %s with empty result or error\n", fullpath)
			} else {
				ret = append(ret, rfile...)
			}
		} else {
			match, err := globber.IsPathMatch(filter, fullpath)
			if err != nil {
				fmt.Printf("Error matching filter %s, to file %s. Got %v", filter, fullpath, err)
			} else if match {
				ret = append(ret, fullpath)
			}
		}
	}
	return ret, nil
}
