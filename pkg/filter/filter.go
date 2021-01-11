package yfilter

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"strconv"
	"strings"

	youtput "github.com/fhquthpdw/yfind/pkg/output"
)

////// Filter Config /////
// filterCfg
type FilterCfg struct {
	path            string
	fileSizeGreater int64
	fileSizeLess    int64
	fileType        map[string]struct{}
	fileName        string
	fileContent     string
	caseSensitive   bool
}

// GetFilterCfg
func NewFilterCfg(fileSizeGreater, fileSizeLess, fileType, fileName, fileContent string, caseSensitive bool) *FilterCfg {
	f := &FilterCfg{}
	f.setFileSizeGreater(fileSizeGreater).
		setFileSizeLess(fileSizeLess).
		setFileType(fileType).
		setFileName(fileName).
		setFileContent(fileContent).
		setCaseSensitive(caseSensitive)
	return f
}

// parseFileSize
func (c *FilterCfg) parseFileSize(size string) int64 {
	n, err := strconv.ParseInt(size[:len(size)-1], 10, 64)
	if err != nil {
		log.Fatalf("invalid limit filesize, parse error: %s", err)
	}

	var limit int64
	switch strings.ToLower(size[len(size)-1:]) {
	case "k":
		limit = 1024 * n
	case "m":
		limit = 1024 * 1024 * n
	case "g":
		limit = 1024 * 1024 * 1024 * n
	default:
		log.Fatalf("invalid limit filesize: %s", size)
	}
	return limit
}

// setFileSizeLess
func (c *FilterCfg) setFileSizeLess(fileSizeLess string) *FilterCfg {
	if fileSizeLess == "" {
		c.fileSizeLess = 0
		return c
	}
	c.fileSizeLess = c.parseFileSize(fileSizeLess)
	return c
}

// setFileSizeGreater
func (c *FilterCfg) setFileSizeGreater(fileSizeGreater string) *FilterCfg {
	if fileSizeGreater == "" {
		c.fileSizeGreater = 0
		return c
	}
	c.fileSizeGreater = c.parseFileSize(fileSizeGreater)
	return c
}

// setFileType
func (c *FilterCfg) setFileType(fileType string) *FilterCfg {
	if fileType == "" {
		return c
	}

	noSpaceTypeStr := strings.ReplaceAll(fileType, " ", "")
	typeSlice := strings.Split(noSpaceTypeStr, ",")
	limitType := make(map[string]struct{}, len(typeSlice))
	for _, v := range typeSlice {
		limitType[v] = struct{}{}
	}
	c.fileType = limitType
	return c
}

// setFileName
func (c *FilterCfg) setFileName(fileName string) *FilterCfg {
	c.fileName = fileName
	return c
}

// setFileContent
func (c *FilterCfg) setFileContent(fileContent string) *FilterCfg {
	c.fileContent = fileContent
	return c
}

// setCaseSensitive
func (c *FilterCfg) setCaseSensitive(cC bool) *FilterCfg {
	c.caseSensitive = cC
	return c
}

///// Filter /////
func NewFilter(cfg *FilterCfg) *Filter {
	f := Filter{
		Cfg: cfg,
	}
	return f.init()
}

type filterFunc func(info os.FileInfo, xargs string) os.FileInfo

type Filter struct {
	Cfg        *FilterCfg
	FilterFuns []filterFunc
}

// init filter functions but not include filterFileContent
func (f *Filter) init() *Filter {
	return f.addFilterFun(f.filterFileSizeGreater).
		addFilterFun(f.filterFileSizeLess).
		addFilterFun(f.filterFileType).
		addFilterFun(f.filterFileName)
}

// DoFilter do filter
// do all filters which register in init function
// and then do filter content function
func (f *Filter) DoFilter(file os.FileInfo, path string) (p bool, o youtput.FileItem) {
	for _, fun := range f.FilterFuns {
		if fun(file, path) == nil {
			return
		}
	}

	// filter file content
	cf, o := f.filterFileContent(file, path)
	if cf == nil {
		return
	}

	return true, o
}

// addFilterFun
func (f *Filter) addFilterFun(fun filterFunc) *Filter {
	f.FilterFuns = append(f.FilterFuns, fun)
	return f
}

// filterFileSizeGreater
func (f *Filter) filterFileSizeGreater(file os.FileInfo, _ string) os.FileInfo {
	if f.Cfg.fileSizeGreater == 0 {
		return file
	}
	if file.Size() >= f.Cfg.fileSizeGreater {
		return file
	}
	return nil
}

// filterFileSizeLess
func (f *Filter) filterFileSizeLess(file os.FileInfo, _ string) os.FileInfo {
	if f.Cfg.fileSizeLess == 0 {
		return file
	}
	if file.Size() <= f.Cfg.fileSizeLess {
		return file
	}
	return nil
}

// filterFileType
func (f *Filter) filterFileType(file os.FileInfo, _ string) os.FileInfo {
	if f.Cfg.fileType == nil {
		return file
	}
	fileSgArr := strings.Split(file.Name(), ".")
	ext := fileSgArr[len(fileSgArr)-1]
	if _, ok := f.Cfg.fileType[ext]; !ok {
		return nil
	}
	return file
}

// filterFileName
func (f *Filter) filterFileName(file os.FileInfo, baseDir string) os.FileInfo {
	if f.Cfg.fileName == "" {
		return file
	}
	fileFullPath := baseDir + file.Name()
	fileNameFilter := f.Cfg.fileName
	// TODO: case sensitive
	// BUG: display all lowercase
	/*
		if !f.Cfg.caseSensitive {
			fileFullPath = strings.ToLower(fileFullPath)
			fileNameFilter = strings.ToLower(fileNameFilter)
		}
	*/

	if strings.Contains(fileFullPath, fileNameFilter) {
		return file
	}

	return nil
}

// filterFileContent
func (f *Filter) filterFileContent(file os.FileInfo, baseDir string) (os.FileInfo, youtput.FileItem) {
	output := youtput.FileItem{}
	fileFullPath := baseDir + file.Name()
	output.FileName = fileFullPath
	output.FileSize = file.Size()

	if f.Cfg.fileContent == "" {
		return file, output
	}

	rFile, err := os.Open(fileFullPath)
	if err != nil {
		// TODO: check if the file is text plain file and check have read permission
		// TODO: if no permission then bring the error message to front
	}
	if rFile == nil {
		//
	}
	defer rFile.Close()

	filterContentStr := f.Cfg.fileContent
	filterContentByte := []byte(filterContentStr)
	var lineNum int64

	scanner := bufio.NewScanner(rFile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lineNum++
		//content := scanner.Text()
		content := scanner.Bytes()
		// TODO: case sensitive
		// BUG: display all lowercase
		if bytes.Contains(content, filterContentByte) {
			//if strings.Contains(content, f.Cfg.fileContent) {
			lineItem := youtput.FileItemLine{Line: lineNum, Content: string(content), Hit: true}
			output.Lines = append(output.Lines, lineItem)
		}
	}

	if len(output.Lines) > 0 {
		return file, output
	}

	return nil, output
}
