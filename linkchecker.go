package linkchecker

import (
	"bytes"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

type Link struct {
	Status int
	URL    string
}

var Links []Link
var Domain string
var checkedLinks []string

func GetPageStatus(page string, client *http.Client) (int, error) {
	resp, err := client.Get(page)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func GrabLinks(doc string) ([]string, error) {
	parsedDoc, err := html.Parse(strings.NewReader(doc))

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
		return []string{}, err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	buf.ReadFrom(resp.Body)

	links, err := GrabLinks(buf.String())
	if err != nil {
		return []string{}, err
	}
	return links, nil
}

func CheckLinks(url string, client *http.Client) error {
	if !isChecked(url) {
		//fmt.Println(url + " wasn't already checked")
		status, err := GetPageStatus(url, client)
		if err != nil {
			return err
		}

		Links = append(Links, Link{status, url})
		checkedLinks = append(checkedLinks, url)
		pageLinks, err := GrabLinksFromServer(url, client)
		if err != nil {
			return err
		}
		for _, l := range pageLinks {
			// TODO improve the URL parsing.
			checkURL := Domain + "/" + l
			CheckLinks(checkURL, client)
		}

	}
	return nil
}

func isChecked(url string) bool {
	for _, i := range checkedLinks {
		if i == url {
			return true
		}
	}
	return false
}
