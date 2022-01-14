package linkchecker

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func RunCLI() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s URL\n", os.Args[0])
		os.Exit(1)
	}
	json := flag.Bool("j", false, "output JSON")
	flag.Parse()

	lc, err := New(*json)
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
