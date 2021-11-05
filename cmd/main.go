package main

import (
	"fmt"
	"linkchecker"
	"log"
	"net/http"
	"os"
)

func main() {

	lc, err := linkchecker.New(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	err = lc.CheckLinks(os.Args[1], http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}

	var lj linkchecker.LinksJSON
	var lt linkchecker.LinksTerminal
	var output string
	if os.Args[2] == "j" {
		output, err = displayLinks(lj, lc)
	} else {
		output, err = displayLinks(lt, lc)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(output)
}

func displayLinks(f linkchecker.Formatter, l *linkchecker.LinkChecker) (string, error) {
	return f.Format(l)
}
