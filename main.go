package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// To accept multiple option string variables
type strslice []string

func (s *strslice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *strslice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

// Variables for CLI options for input
var (
	urls     strslice     //multiple input with -url=xxx -url=yyy
	sword    string       //search word to be replaced with, -convert-search-word=xxx
	rword    string       //replace word, -convert-replace-word=yyy
	outfile  string       //output file, -out-file=xxx. If this is not specified, Output is stdout.
	exitCode int      = 0 //total CLI exitCode
)

// Interfaces
// Loader interface
type Loader interface {
	Load() ([]string, error)
}

// Converter interface
type Converter interface {
	Convert(*[]string) []string
}

// Printer interface
type Printer interface {
	Print(*[]string)
}

// URLLoader class implements Loader
type URLLoader struct {
	urls *[]string
}

func (u *URLLoader) Load() ([]string, error) {
	var bodies []string = make([]string, 0)
	var resp *http.Response
	var err error
	var body []byte
	for _, url := range *u.urls {
		resp, err = http.Get(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to do http request: %s", err.Error())
			exitCode = 1
			return bodies, err
		}
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read from http body: %s", err.Error())
			exitCode = 2
			return bodies, err
		}
		bodies = append(bodies, string(body))
	}
	defer resp.Body.Close()

	return bodies, nil
}

func NewUrlLoader(urls *[]string) Loader {
	return &URLLoader{
		urls: urls,
	}
}

// ReplaceConverter class implements Converter
type ReplaceConverter struct {
	search  string
	replace string
}

func (rep *ReplaceConverter) Convert(rss *[]string) []string {
	var res []string = make([]string, 0)
	for _, r := range *rss {
		res = append(res, strings.Replace(r, rep.search, rep.replace, -1))
	}

	return res
}

func NewReplaceConverter(search string, replace string) Converter {
	return &ReplaceConverter{
		search:  search,
		replace: replace,
	}
}

// OutputPrinter class implements Printer
type OutputPrinter struct {
}

func (op *OutputPrinter) Print(rss *[]string) {
	for _, r := range *rss {
		fmt.Println(r)
	}

	return
}

func NewOutputPrinter() Printer {
	return &OutputPrinter{}
}

// FileOutputPrinter class implements Printer
type FileOutputPrinter struct {
	outputfile string
}

func (fop *FileOutputPrinter) Print(rss *[]string) {
	var fpw *os.File
	var err error
	fpw, err = os.Create(fop.outputfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open output file: %s", err.Error())
		exitCode = 3
		return
	}
	defer fpw.Close()

	var writer *bufio.Writer = bufio.NewWriter(fpw)
	for _, r := range *rss {
		fmt.Fprint(writer, r)
		writer.Flush()
	}
}

func NewFileOutputPrinter(outputfile string) Printer {
	return &FileOutputPrinter{
		outputfile: outputfile,
	}
}

// RSSDocument class
type RSSDocument struct {
	rawrss    []string
	loader    Loader
	converter Converter
	printer   Printer
}

func (rd *RSSDocument) LoadRSS() {
	var err error
	rd.rawrss, err = rd.loader.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load RSS")
	}
	return
}

func (rd *RSSDocument) ConvertRSS() {
	rd.rawrss = rd.converter.Convert(&rd.rawrss)
	return
}

func (rd *RSSDocument) PrintRSS() {
	rd.printer.Print(&rd.rawrss)
	return
}

func NewRSSDocument(urls *[]string, sword string, rword string, outfile string) *RSSDocument {
	var printer Printer
	if outfile == "" {
		printer = NewOutputPrinter()
	} else {
		printer = NewFileOutputPrinter(outfile)
	}
	var loader Loader = NewUrlLoader(urls)
	var converter Converter = NewReplaceConverter(sword, rword)

	return &RSSDocument{
		rawrss:    make([]string, 0),
		loader:    loader,
		converter: converter,
		printer:   printer,
	}
}

// Initialize only once at execution
func init() {
	// Parse CLI flags
	flag.Var(&urls, "url", "URL to input RSS")
	flag.StringVar(&sword, "convert-search-word", "", "Word to be replaced with")
	flag.StringVar(&rword, "convert-replace-word", "", "Word to replace with")
	flag.StringVar(&outfile, "out-file", "", "Output file path")
	flag.Parse()

	if len(urls) <= 0 {
		fmt.Fprintf(os.Stderr, "Need to specify 1 -url option at least.")
		os.Exit(1)
	}
}

func main() {
	fmt.Println("This is rssconv!")
	fmt.Printf("%v\n", urls)
	fmt.Printf("%v\n", sword)
	fmt.Printf("%v\n", rword)
	fmt.Printf("%v\n", outfile)

	var cnvurls []string = urls
	var rssdoc *RSSDocument = NewRSSDocument(&cnvurls, sword, rword, outfile)
	rssdoc.LoadRSS()
	rssdoc.ConvertRSS()
	rssdoc.PrintRSS()

	os.Exit(exitCode)

}
