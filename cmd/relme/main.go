// Command relme will retrieve all rel="me" links from a URL, and optionally
// verify that they also point back to the URL.
package main

import (
	"flag"
	"fmt"
	"log"

	"hawx.me/code/relme"
)

func printHelp() {
	fmt.Println(`Usage: relme [-verify] URL`)
}

func run(url string, verify bool) ([]string, error) {
	links, err := relme.Find(url)

	if verify && err == nil {
		links, err = relme.Verify(url, links)
	}

	return links, err
}

func main() {
	verify := flag.Bool("verify", false, "")

	flag.Usage = func() { printHelp() }
	flag.Parse()

	if flag.NArg() == 0 {
		printHelp()
		return
	}

	url := flag.Arg(0)

	links, err := run(url, *verify)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, link := range links {
		fmt.Println(link)
	}
}
