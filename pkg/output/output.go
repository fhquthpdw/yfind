package youtput

import (
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
)

func NewOutput(filterFileName, filterFileContent string) *Output {
	return &Output{
		FilterFileName:    filterFileName,
		FilterFileContent: filterFileContent,
	}
}

type FileItemLine struct {
	Line    int64
	Content string
	Hit     bool
}

type FileItem struct {
	FileName string
	FileSize int64
	Lines    []FileItemLine
}

type Output struct {
	FilterFileName    string
	FilterFileContent string
}

func (o *Output) Output(wg *sync.WaitGroup, fileItemChan chan FileItem) {
	defer wg.Done()

	looping := true
	for looping {
		select {
		case fileItem := <-fileItemChan:
			if fileItem.FileName == "" && len(fileItem.Lines) == 0 {
				looping = false
			} else {
				o.colorOutput(fileItem)
			}
		}
	}

	return
}

func (o *Output) colorOutput(fileItem FileItem) {
	clFileName := color.New(color.FgCyan)
	oclFileName := color.New(color.FgGreen)
	o.printFileName(fileItem, clFileName, oclFileName)

	clLine := color.New(color.FgRed, color.Bold, color.Italic)
	oclLine := color.New()
	o.printLines(fileItem, clLine, oclLine)

	if o.FilterFileContent != "" {
		fmt.Println("=======================================")
		fmt.Println()
	}
}

func (o *Output) printFileName(fileItem FileItem, cl *color.Color, ocl *color.Color) {
	cl.Print(">>> ")
	cl.Print(o.formatOutputSize(fileItem.FileSize), " ")
	if o.FilterFileName != "" {
		o.colorTextInLine(fileItem.FileName, o.FilterFileName, cl, ocl)
	} else {
		_, _ = ocl.Println(fileItem.FileName)
	}
}

func (o *Output) printLines(fileItem FileItem, cl *color.Color, ocl *color.Color) {
	if len(fileItem.Lines) == 0 {
		return
	}
	if o.FilterFileContent == "" {
		return
	}
	lineNumColor := color.New(color.FgBlue)
	for _, l := range fileItem.Lines {
		_, _ = lineNumColor.Print(l.Line)
		o.colorTextInLine(l.Content, o.FilterFileContent, cl, ocl)
	}
	return
}

func (o *Output) colorTextInLine(lineText, colorText string, cl *color.Color, ocl *color.Color) {
	strArr := strings.Split(lineText, colorText)
	for idx, t := range strArr {
		_, _ = ocl.Print(t)
		if idx < len(strArr)-1 {
			_, _ = cl.Print(colorText)
		}
	}
	fmt.Println()
}

func (o *Output) formatOutputSize(sizeByte int64) string {
	const (
		KB = 1024
		MB = 1024 * 1024
		GB = 1024 * 1024 * 1024
	)

	sizeByteFloat := float64(sizeByte)
	if sizeByte < KB {
		return fmt.Sprintf("%dB", sizeByte)
	} else {
		if sizeByte >= KB && sizeByte < MB {
			return fmt.Sprintf("%.2fK", sizeByteFloat/1024)
		} else if sizeByte >= MB && sizeByte < GB {
			return fmt.Sprintf("%.2fM", sizeByteFloat/1024/1024)
		} else {
			return fmt.Sprintf("%.2fG", sizeByteFloat/1024/1024/1024)
		}
	}
	return ""
}
