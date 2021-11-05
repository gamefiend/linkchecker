package linkchecker

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type Link struct {
	Status int
	URL    string
}

type LinkChecker struct {
	Domain       string
	CheckedLinks []string
	Links        []Link
	CheckCurrent int
	CheckLimit   int
	Debug        bool
}

func New(URL string) (*LinkChecker, error) {
	domain, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	return &LinkChecker{
		Domain:       domain.Host,
		CheckedLinks: []string{},
		Links:        []Link{},
		CheckCurrent: 0,
		CheckLimit:   4,
		Debug:        false,
	}, nil
}
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

func (lc *LinkChecker) CheckLinks(URL string, client *http.Client) error {
	lc.debug("Check links called with ", URL)

	if lc.CheckCurrent >= lc.CheckLimit {
		lc.debug("Hit Check limit of", lc.CheckCurrent)
		return nil
	}
	lc.debug("CheckCurrent ", lc.CheckCurrent)
	if lc.isChecked(URL) {
		lc.debug("Skipping ", URL, " already checked")
		return nil
	}

	status, err := GetPageStatus(URL, client)
	if err != nil {
		return err
	}

	lc.Links = append(lc.Links, Link{status, URL})

	lc.CheckedLinks = append(lc.CheckedLinks, URL)
	pageLinks, err := GrabLinksFromServer(URL, client)
	if err != nil {
		return err
	}
	for _, l := range pageLinks {
		// TODO improve the URL parsing.
		var checkURL string
		test, _ := url.Parse(l)
		if test.IsAbs() {
			checkURL = l
		} else {
			checkURL = "https://" + lc.Domain + "/" + l
		}

		lc.CheckLinks(checkURL, client)
	}
	lc.debug("Check ", lc.CheckCurrent)
	lc.CheckCurrent++
	return nil
}

func (lc LinkChecker) isChecked(URL string) bool {
	for _, i := range lc.CheckedLinks {
		if i == URL {
			return true
		}
	}
	return false
}

func (lc *LinkChecker) debug(args ...interface{}) {
	if lc.Debug {
		fmt.Printf("%v", args)
	}
}
