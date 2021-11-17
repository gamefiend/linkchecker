package main

import (
	"flag"
	"fmt"
	"linkchecker"
	"log"
	"net/http"
)

func main() {
	var format, site string

	flag.StringVar(&format, "format", "terminal", "format for results to be returned in (json or terminal)")
	flag.StringVar(&site, "site", "", "site to check links from")
	flag.Parse()

	lc, err := linkchecker.New(site)
	if err != nil {
		log.Fatal(err)
	}
	err = lc.CheckLinks(site, http.DefaultClient)
	lc.Workers.Wait()
	if err != nil {
		log.Fatal(err)
	}

	// Output
	var output string
	switch format {
	case "json":
		output, err = displayLinks(linkchecker.LinksJSON{}, lc)
	default:
		output, err = displayLinks(linkchecker.LinksTerminal{}, lc)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(output)
}

func displayLinks(f linkchecker.Formatter, l *linkchecker.LinkChecker) (string, error) {
	return f.Format(l)
}
