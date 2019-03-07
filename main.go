package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/gadelkareem/go-helpers"
	"io"
	"os"
	"strings"
)

type urlErr struct {
	u   string
	err string
}

var (
	brokenUrls []urlErr
	infile     = flag.String("infile", "urls.csv", "urls file path")
	outfile    = flag.String("outfile", "results.csv", "results file path")
	comma      = flag.String("comma", ";", "splitter character")
	urlCol     = flag.Int("urlCol", 1, "URL column")
)

func main() {
	csvFile, err := os.Open(*infile)
	h.PanicOnError(err)
	defer csvFile.Close()
	r := csv.NewReader(csvFile)
	r.Comma = []rune(*comma)[0]
	r.Read() //header
loop:
	for {
		line, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		u := line[*urlCol]
		fmt.Printf("URL: %s ", u)
		content, err := h.GetUrl(u)
		if err != nil {
			addError("Broken URL", u, err)
			continue
		}
		if content == "" {
			addError("Empty page", u, nil)
			continue
		}

		for i := *urlCol + 1; i < len(line); i++ {
			k := strings.TrimSpace(line[i])
			if !strings.Contains(content, k) {
				addError(fmt.Sprintf("Keyword %s not found", k), u, nil)
				continue loop
			}
		}
		fmt.Printf("✅ works!\n")
	}
	file, err := os.Create(*outfile)
	h.PanicOnError(err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{"URL", "Error"})
	h.PanicOnError(err)
	for _, v := range brokenUrls {
		err = writer.Write([]string{v.u, v.err})
		h.PanicOnError(err)
	}

}

func addError(report string, u string, err error) {
	brokenUrls = append(brokenUrls, urlErr{u: u, err: report})
	fmt.Printf("❌ Error: %s\n", report)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
