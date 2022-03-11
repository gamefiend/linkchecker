package linkchecker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"golang.org/x/net/html"
)

type LinkStatus string

const (
	LinkStatusOK       LinkStatus = "OK"
	LinkStatusWarning             = "Warning"
	LinkStatusCritical            = "Critical"
)

type Result struct {
	LinkStatus LinkStatus
	HTTPStatus int
	Link       string
}

func (r Result) String() string {
	switch r.LinkStatus {
	case LinkStatusOK:
		return fmt.Sprintf("%s %s", r.Link, color.GreenString(string(r.LinkStatus)))
	case LinkStatusWarning:

		return fmt.Sprintf("%s %s", r.Link, color.YellowString(string(r.LinkStatus)))
	case LinkStatusCritical:

		return fmt.Sprintf("%s %s", r.Link, color.RedString(string(r.LinkStatus)))
	default:
		return fmt.Sprintf("%s %s", r.Link, r.LinkStatus)
	}

}

func (r Result) ToJSON() string {
	j, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return string(j)
}

type LinkChecker struct {
	Domain       string
	CheckedLinks []string
	CheckCurrent int
	CheckLimit   int
	Debug        bool
	Workers      sync.WaitGroup
	HTTPClient   *http.Client
	stream       chan Result
	jsonMode     bool
	verboseMode  bool
}

type option func(*LinkChecker) error

func New(options ...option) (*LinkChecker, error) {
	lc := &LinkChecker{
		CheckedLinks: []string{},
		CheckCurrent: 0,
		CheckLimit:   4,
		Debug:        false,
		Workers:      sync.WaitGroup{},
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		stream:      make(chan Result, 10),
		jsonMode:    false,
		verboseMode: false,
	}
	for _, opt := range options {
		err := opt(lc)
		if err != nil {
			return nil, err
		}
	}
	return lc, nil
}

func WithJSONOutput() option {
	return func(lc *LinkChecker) error {
		lc.jsonMode = true
		return nil
	}
}

func WithVerboseOutput() option {
	return func(lc *LinkChecker) error {
		lc.verboseMode = true
		return nil
	}
}

func (lc *LinkChecker) GetHTTPStatus(page string) (int, error) {
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

	// Adding an additional Worker for the initial run of CheckLinks
	lc.Workers.Add(1)
	go func() {
		lc.Workers.Wait()
		close(lc.stream)
	}()
	return lc.CheckLinks(link)
}

func (lc *LinkChecker) CheckLinks(link string) error {
	// we've already incremented workers in Check
	defer lc.Workers.Done()
	lc.debug("Check links called with ", link)
	link = lc.CanonicaliseLink(link)
	_, err := url.Parse(link)
	if err != nil {
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
	httpStatus, err := lc.GetHTTPStatus(link)
	if err != nil {
		return err
	}

	result := Result{
		LinkStatus: MapHTTPToLink(httpStatus),
		HTTPStatus: httpStatus,
		Link:       link,
	}

	if lc.verboseMode || result.LinkStatus != LinkStatusOK {
		lc.stream <- result
	}
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

func (lc LinkChecker) StreamResults() <-chan Result {
	return lc.stream
}

func (lc LinkChecker) AllResults() []Result {
	var result []Result
	for r := range lc.StreamResults() {
		result = append(result, r)
	}
	return result
}

func MapHTTPToLink(httpStatus int) LinkStatus {
	switch httpStatus {
	case http.StatusOK:
		return LinkStatusOK
	case http.StatusInternalServerError:
		return LinkStatusWarning
	default:
		return LinkStatusCritical
	}
}
