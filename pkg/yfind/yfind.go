package yfind

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	yfilter "github.com/fhquthpdw/yfind/pkg/filter"

	youtput "github.com/fhquthpdw/yfind/pkg/output"
)

func NewYFind(filter *yfilter.Filter, output *youtput.Output) *Yfind {
	return &Yfind{
		Filter: filter,
		Output: output,
	}
}

type Yfind struct {
	RootPath string
	Filter   *yfilter.Filter
	Output   *youtput.Output
}

type FileItem youtput.FileItem

func (f *Yfind) SetRootPath(path string) *Yfind {
	if path == "" {
		curPath, err := os.Getwd()
		if err != nil {
			log.Fatalf(err.Error())
		}
		path = curPath
	}
	f.RootPath = path
	return f
}

func (f *Yfind) timeCostTrace(t time.Time) {
	fmt.Println("Time Cost: ", time.Since(t))
}

func (f *Yfind) Run() {
	defer f.timeCostTrace(time.Now())

	var wg sync.WaitGroup
	wg.Add(2)

	outputChan := make(chan youtput.FileItem, 10)
	// scan files, do filter, write filtered data to channel
	go func(wg *sync.WaitGroup, outputChan chan youtput.FileItem) {
		f.workDir(f.RootPath, wg, outputChan)
		close(outputChan)
	}(&wg, outputChan)

	// get channel data and output to console
	go f.Output.Output(&wg, outputChan)

	// TODO: output scanned total dirs, files, result files, lines
	// TODO: show time cost here
	// TODO: catch os.Single
	wg.Wait()
}

func (f *Yfind) workDir(path string, wg *sync.WaitGroup, outputChan chan youtput.FileItem) {
	defer wg.Done()

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("%s: %s\n", path, err)
	}

	path = strings.TrimRight(path, "/") + "/"

	var filterContentWg sync.WaitGroup
	for _, file := range files {
		fName := path + file.Name()

		// work dir
		if file.IsDir() {
			wg.Add(1)
			f.workDir(fName, wg, outputChan)
			continue
		}

		// work file
		if f.Output.FilterFileContent == "" { // no content filter, no more goroutines
			if pass, o := f.workFile(file, path); pass {
				outputChan <- o
			}
		} else { // goroutines working on content filter
			filterContentWg.Add(1)
			go func(file os.FileInfo, path string, wg *sync.WaitGroup) {
				defer wg.Done()

				if pass, o := f.workFile(file, path); pass {
					outputChan <- o
				}
			}(file, path, &filterContentWg)
		}
	}
	filterContentWg.Wait()
}

func (f *Yfind) workFile(file os.FileInfo, path string) (p bool, o youtput.FileItem) {
	return f.Filter.DoFilter(file, path)
}
