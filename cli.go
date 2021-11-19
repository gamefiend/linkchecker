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
	lc, err := New()
	if err != nil {
		log.Fatal(err)
	}
	err = lc.Check(os.Args[1])
	lc.Workers.Wait()
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range lc.Results {
		if *json {
			fmt.Println(r.ToJSON())
		} else {
			fmt.Println(r)
		}
	}
}
