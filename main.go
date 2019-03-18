package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/gadelkareem/go-helpers"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

type urlErr struct {
	u    string
	err  string
	line int32
}

const MaxConcurrency = 200

var (
	brokenUrlsMu     sync.Mutex
	brokenUrls       []urlErr
	infile                 = flag.String("infile", "urls.csv", "urls file path")
	outfile                = flag.String("outfile", "results.csv", "results file path")
	comma                  = flag.String("comma", ";", "splitter character")
	urlCol                 = flag.Int("urlCol", 0, "URL column")
	total, broken, l int32 = 0, 0, 0
)
// go build ; ./bulk-url-checker -infile=fixed.csv
// go build ; ./bulk-url-checker -urlCol=1
func main() {
	flag.Parse()
	h.LiftRLimits()
	csvFile, err := os.Open(*infile)
	h.PanicOnError(err)
	defer csvFile.Close()
	fmt.Printf("Reading file %s\n", *infile)
	r := csv.NewReader(csvFile)
	r.Comma = []rune(*comma)[0]
	r.Read() //header
	wg := h.NewWgExec(MaxConcurrency)

	for {
		l++
		line, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		u := line[*urlCol]
		if u == "" {
			continue
		}
		total++
		wg.Run(check, u, line)

	}
	wg.Wait()
	file, err := os.Create(*outfile)
	h.PanicOnError(err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{"URL", "Error"})
	h.PanicOnError(err)
	fmt.Fprintf(os.Stderr, "\n\nError Report: \n\n")
	for _, v := range brokenUrls {
		err = writer.Write([]string{v.u, v.err})
		fmt.Fprintf(os.Stderr, "URL: %s ❌ Error: %s Line (%d)\n", v.u, v.err, v.line)
		h.PanicOnError(err)
	}

	fmt.Printf("Found %d broken URLs from total %d URLs", broken, total)
}

func addError(report string, u string, err error, lineNumber int32) {
	brokenUrlsMu.Lock()
	brokenUrls = append(brokenUrls, urlErr{u: u, err: report, line: lineNumber})
	brokenUrlsMu.Unlock()
}

func check(param ...interface{}) {
	u := param[0].(string)
	line := param[1].([]string)
	content, err := h.GetUrl(u)
	if err != nil {
		addError("Broken URL", u, err, l)
		atomic.AddInt32(&broken, 1)
		return
	}
	if content == "" {
		addError("Empty page", u, nil, l)
		atomic.AddInt32(&broken, 1)
		return
	}

	for i := *urlCol + 1; i < len(line); i++ {
		k := strings.ToLower(strings.TrimSpace(line[i]))
		content = strings.ToLower(content)
		if !strings.Contains(content, k) {
			addError(fmt.Sprintf("Keyword %s not found", k), u, nil, l)
			atomic.AddInt32(&broken, 1)
			return
		}
	}
	fmt.Printf("URL: %s ✅ works!\n", u)
}
