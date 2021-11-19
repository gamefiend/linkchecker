package linkchecker

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type Result struct {
	Status int
	Link   string
}

type LinkChecker struct {
	Domain       string
	CheckedLinks []string
	Links        []Result
	CheckCurrent int
	CheckLimit   int
	Debug        bool
	Workers      sync.WaitGroup
	HTTPClient   *http.Client
}

func New() (*LinkChecker, error) {
	return &LinkChecker{
		CheckedLinks: []string{},
		Links:        []Result{},
		CheckCurrent: 0,
		CheckLimit:   4,
		Debug:        false,
		Workers:      sync.WaitGroup{},
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (lc *LinkChecker) GetPageStatus(page string) (int, error) {
	resp, err := lc.HTTPClient.Get(page)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func (lc *LinkChecker) Check(link string) error {
	URL, err := url.Parse(link)
	if err != nil {
		return err
	}
	lc.Domain = URL.Host
	return lc.CheckLinks(link)
}

func (lc *LinkChecker) CheckLinks(link string) error {
	lc.debug("Check links called with ", link)
	link = lc.CanonicaliseLink(link)
	_, err := url.Parse(link)
	if err != nil {
		lc.Links = append(lc.Links, Result{
			Link: link,
		})
		return nil
	}
	if lc.CheckCurrent >= lc.CheckLimit {
		lc.debug("Hit Check limit of", lc.CheckCurrent)
		return nil
	}
	lc.debug("CheckCurrent ", lc.CheckCurrent)
	if lc.isChecked(link) {
		lc.debug("Skipping ", link, " already checked")
		return nil
	}
	lc.CheckedLinks = append(lc.CheckedLinks, link)
	status, err := lc.GetPageStatus(link)
	if err != nil {
		return err
	}

	lc.Links = append(lc.Links, Result{status, link})

	lc.CheckedLinks = append(lc.CheckedLinks, link)
	if lc.IsExternal(link) {
		return nil
	}
	pageLinks, err := lc.GrabLinksFromServer(link)
	if err != nil {
		return err
	}
	for _, l := range pageLinks {
		lc.Workers.Add(1)
		go func(l string) {
			defer lc.Workers.Done()
			lc.CheckLinks(l)
		}(l)
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

func (lc LinkChecker) IsExternal(link string) bool {
	startsWithHttpsDomain := strings.HasPrefix(link, "https://"+lc.Domain)
	startsWithHttpDomain := strings.HasPrefix(link, "http://"+lc.Domain)
	return !startsWithHttpDomain && !startsWithHttpsDomain
}

func (lc LinkChecker) CanonicaliseLink(link string) string {
	var scheme, host string
	if !strings.HasPrefix(link, "https://") {
		scheme = "https://"
	}
	if !strings.HasPrefix(link, lc.Domain) && !strings.HasPrefix(link, "https://"+lc.Domain) {
		host = lc.Domain + "/"
	}
	return scheme + host + link
}

func (lc *LinkChecker) debug(args ...interface{}) {
	if lc.Debug {
		fmt.Printf("%v", args)
	}
}

func (lc *LinkChecker) GrabLinksFromServer(url string) ([]string, error) {
	resp, err := lc.HTTPClient.Get(url)
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
