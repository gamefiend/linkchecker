package linkchecker

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

func GetPageStatus(page string, client *http.Client) (int, error) {
	resp, err := client.Get(page)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

func GrabLinks(doc string) ([]string, error) {
	parsedDoc, err := html.Parse(strings.NewReader(doc))
	fmt.Println(doc)
	if err != nil {
		return []string{}, err
	}

	// traverse the parsed html looking for hrefs
	var links []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					fmt.Println(a.Val)
					links = append(links, a.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(parsedDoc)
	return links, nil
}

func GrabLinksFromServer(url string, client *http.Client) ([]string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	page
	links := GrabLinks()
}
