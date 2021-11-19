package linkchecker_test

import (
	"fmt"
	"io"
	"linkchecker"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFetchStatusCodeFromPage(t *testing.T) {
	t.Parallel()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "here is your page")
	}))
	page := ts.URL + "/test"
	lc, err := linkchecker.New()
	lc.HTTPClient = ts.Client()
	status, err := lc.GetPageStatus(page)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Errorf("Wanted %d got %d", http.StatusOK, status)
	}
}

func TestGrabLinksFromPage(t *testing.T) {
	t.Parallel()

	want := []string{"whatever.html", "you.html"}
	file, err := os.ReadFile("testdata/links.html")
	if err != nil {
		t.Fatal(err)
	}
	got, err := linkchecker.GrabLinks(string(file))
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Errorf("want %v, got: %v", want, got)
	}
}

func TestGrabLinksFromServer(t *testing.T) {
	t.Parallel()

	want := []string{"whatever.html", "you.html"}
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, err := os.ReadFile("testdata/links.html")
		if err != nil {
			fmt.Print("error")
		}
		io.Copy(w, strings.NewReader(string(file)))
	}))
	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	lc.HTTPClient = ts.Client()
	got, err := lc.GrabLinksFromServer(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Errorf("want %v, got: %v", want, got)
	}
}

func TestCheckLinksReturnsAllPages(t *testing.T) {
	ts := httptest.NewTLSServer(http.FileServer(http.Dir("testdata")))
	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	lc.HTTPClient = ts.Client()
	want := []linkchecker.Link{
		{
			Status: 200,
			URL:    ts.URL + "/links.html",
		},
		{
			Status: 200,
			URL:    ts.URL + "/whatever.html",
		},
		{
			Status: 404,
			URL:    ts.URL + "/me.html",
		},
		{
			Status: 200,
			URL:    ts.URL + "/you.html",
		},
	}
	startLink := ts.URL + "/links.html"
	err = lc.CheckLinks(startLink)
	lc.Workers.Wait()
	if err != nil {
		t.Fatal(err)
	}
	// since we grab slices out of order, sort the slices so they are comparable
	sort.Slice(want, func(i, j int) bool {
		return want[i].URL < want[j].URL
	})

	got := lc.Links
	sort.Slice(got, func(i, j int) bool {
		return got[i].URL < got[j].URL
	})

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestLinkCheckerNew(t *testing.T) {
	t.Parallel()
	var lc *linkchecker.LinkChecker
	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	_ = lc
}
