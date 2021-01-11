/*
Copyright © 2021 daochun.zhao <daochun.zhao@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/fhquthpdw/yfind/pkg/yfind"

	youtput "github.com/fhquthpdw/yfind/pkg/output"

	yfilter "github.com/fhquthpdw/yfind/pkg/filter"

	"github.com/spf13/cobra"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yfind",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: Run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ffind.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//
	rootCmd.PersistentFlags().StringVar(&path, "path", "", "root path")
	rootCmd.PersistentFlags().StringVar(&fileSizeGreater, "size-greater", "", "limit file size greater: 1k|2m|3g")
	rootCmd.PersistentFlags().StringVar(&fileSizeLess, "size-less", "", "limit file size less: 1k|2m|3g")
	rootCmd.PersistentFlags().StringVar(&fileType, "type", "", "limit file type: txt,go")
	rootCmd.PersistentFlags().StringVar(&fileName, "name", "", "search file name")
	rootCmd.PersistentFlags().StringVar(&fileContent, "content", "", "search file content")
	rootCmd.PersistentFlags().BoolVar(&cC, "no-cc", true, "case sensitive")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".yfind" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".yfind")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

///// YFind Run /////
var (
	path            string
	fileSizeGreater string
	fileSizeLess    string
	fileType        string
	fileName        string
	fileContent     string
	cC              bool
)

//func Run(cmd *cobra.Command, args []string) {
func Run(_ *cobra.Command, _ []string) {
	// pprof
	fCpu, _ := os.Create("./cpu.pprof")
	fMem, _ := os.Create("./mem.pprof")
	fGR, _ := os.Create("./goroutine.pprof")
	pprof.StartCPUProfile(fCpu)
	pprof.WriteHeapProfile(fMem)
	pprof.Lookup("goroutine").WriteTo(fGR, 1)
	defer func() {
		pprof.StopCPUProfile()
		fCpu.Close()
		fMem.Close()
		fGR.Close()
	}()

	yFilter := yfilter.NewFilter(yfilter.NewFilterCfg(fileSizeGreater, fileSizeLess, fileType, fileName, fileContent, cC))
	yOutput := youtput.NewOutput(fileName, fileContent)
	yFind := yfind.NewYFind(yFilter, yOutput)
	yFind.SetRootPath(path).Run()
}

/*
///// FFind Config /////
// NewFFindCfg
func NewFFindCfg() *ffindCfg {
	return &ffindCfg{}
}

// ffindCfg
type ffindCfg struct {
	path            string
	fileSizeGreater int64
	fileSizeLess    int64
	fileType        map[string]struct{}
	fileName        string
	fileContent     string
	caseSensitive   bool
}

// GetCfg
func (c *ffindCfg) GetCfg() *ffindCfg {
	return c.
		setPath().
		setFileSizeGreater().
		setFileSizeLess().
		setFiletype().
		setFileName().
		setFileContent().
		setCaseSensitive()
}

// setPath set scanner root path
// if not set than get ffind current path
func (c *ffindCfg) setPath() *ffindCfg {
	if path == "" {
		curPath, err := os.Getwd()
		if err != nil {
			log.Fatalf(err.Error())
		}
		path = curPath
	}
	c.path = path
	return c
}

func (c *ffindCfg) parseFileSize(size string) int64 {
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

func (c *ffindCfg) setFileSizeLess() *ffindCfg {
	if fileSizeLess == "" {
		c.fileSizeLess = 0
		return c
	}
	c.fileSizeLess = c.parseFileSize(fileSizeLess)
	return c
}

func (c *ffindCfg) setFileSizeGreater() *ffindCfg {
	if fileSizeGreater == "" {
		c.fileSizeGreater = 0
		return c
	}
	c.fileSizeGreater = c.parseFileSize(fileSizeGreater)
	return c
}

func (c *ffindCfg) setFiletype() *ffindCfg {
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
func (c *ffindCfg) setFileName() *ffindCfg {
	c.fileName = fileName
	return c
}
func (c *ffindCfg) setFileContent() *ffindCfg {
	c.fileContent = fileContent
	return c
}
func (c *ffindCfg) setCaseSensitive() *ffindCfg {
	c.caseSensitive = cC
	return c
}

///// FFind Config /////

///// FFind /////
func NewFFind(conf *ffindCfg) *ffind {
	f := &ffind{
		conf: conf,
	}
	return f
}

type ffind struct {
	conf      *ffindCfg
	filterFun []func(os.FileInfo) os.FileInfo
}

func (f *ffind) Run() {
	f.run()
}

type output struct {
	fileName string
	lines    []struct {
		line    int64
		content string
	}
}

func (f *ffind) run() {
	f.init()

	outputChan := make(chan output, 10)

	var wg sync.WaitGroup
	wg.Add(2)
	go func(wg *sync.WaitGroup, outputChan chan output) {
		f.workDir(f.conf.path, wg, outputChan)
		close(outputChan)
	}(&wg, outputChan)
	go f.output(&wg, outputChan)
	wg.Wait()
}

func (f *ffind) workDir(path string, wg *sync.WaitGroup, outputChan chan output) {
	defer wg.Done()

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("%s: %s\n", path, err)
	}

	path = strings.TrimRight(path, "/") + "/"

	for _, file := range files {
		fName := path + file.Name()

		// work dir
		if file.IsDir() {
			wg.Add(1)
			f.workDir(fName, wg, outputChan)
			continue
		}

		// work file
		pass := f.workFile(file)
		if pass {
			outputChan <- output{
				fileName: fName,
			}
		}
	}
}

func (f *ffind) workFile(file os.FileInfo) (pass bool) {
	pass = true
	for _, fun := range f.filterFun {
		if fun(file) == nil {
			pass = false
			break
		}
	}
	return pass
}

func (f *ffind) output(wg *sync.WaitGroup, outputChan chan output) {
	defer wg.Done()

	looping := true
	for looping {
		select {
		case outputItem := <-outputChan:
			if outputItem.fileName == "" && len(outputItem.lines) == 0 {
				looping = false
			} else {
				f.colorOutput(outputItem)
			}
		}
	}

	return
}

func (f *ffind) colorOutput(output output) {
	// output file name
	fileName := output.fileName
	if f.conf.fileName != "" {
		cl := color.New(color.FgCyan).Add(color.Underline)
		fNameArr := strings.Split(fileName, f.conf.fileName)
		for idx, t := range fNameArr {
			fmt.Print(t)
			if idx < len(fNameArr)-1 {
				cl.Print(f.conf.fileName)
			}
		}
		fmt.Println()
	} else {
		fmt.Println(fileName)
	}

	// output file lines
}

func (f *ffind) init() *ffind {
	return f.
		addFilterFun(f.filterFileSizeGreater).
		addFilterFun(f.filterFileSizeLess).
		addFilterFun(f.filterFileType).
		addFilterFun(f.filterFileName).
		addFilterFun(f.filterFileContent)
}

func (f *ffind) addFilterFun(fun func(info os.FileInfo) os.FileInfo) *ffind {
	f.filterFun = append(f.filterFun, fun)
	return f
}

func (f *ffind) filterFileSizeGreater(file os.FileInfo) os.FileInfo {
	if f.conf.fileSizeGreater == 0 {
		return file
	}
	if file.Size() >= f.conf.fileSizeGreater {
		return file
	}
	return nil
}

func (f *ffind) filterFileSizeLess(file os.FileInfo) os.FileInfo {
	if f.conf.fileSizeLess == 0 {
		return file
	}
	if file.Size() <= f.conf.fileSizeLess {
		return file
	}
	return nil
}

func (f *ffind) filterFileType(file os.FileInfo) os.FileInfo {
	if f.conf.fileType == nil {
		return file
	}
	fileSgArr := strings.Split(file.Name(), ".")
	ext := fileSgArr[len(fileSgArr)-1]
	if _, ok := f.conf.fileType[ext]; !ok {
		return nil
	}
	return file
}

func (f *ffind) filterFileName(file os.FileInfo) os.FileInfo {
	if f.conf.fileName == "" {
		return file
	}
	if strings.Contains(file.Name(), f.conf.fileName) {
		return file
	}

	return nil
}

func (f *ffind) filterFileContent(file os.FileInfo) os.FileInfo {
	if f.conf.fileContent == "" {
		return file
	}
	return file
}
*/
