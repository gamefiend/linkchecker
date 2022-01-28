package linkchecker

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func RunCLI() {
	options := []option{}
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s URL\n", os.Args[0])
		os.Exit(1)
	}

	json := flag.Bool("j", false, "output JSON")
	verbose := flag.Bool("v", false, "verbose output")

	flag.Parse()
	if *json {
		options = append(options, WithJSONOutput())
	}
	if *verbose {
		options = append(options, WithVerboseOutput())
	}
	lc, err := New(options...)
	if err != nil {
		log.Fatal(err)
	}
	args := flag.Args()
	err = lc.Check(args[0])
	if err != nil {
		log.Fatal(err)
	}
	for r := range lc.StreamResults() {

		fmt.Println(r)
	}

}
