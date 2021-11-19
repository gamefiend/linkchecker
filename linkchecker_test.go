package linkchecker_test

import (
	"fmt"
	"io"
	"linkchecker"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestFetchStatusCodeFromPage(t *testing.T) {
	t.Parallel()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "here is your page")
	}))
	page := ts.URL + "/test"
	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	lc.HTTPClient = ts.Client()
	status, err := lc.GetPageStatus(page)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Errorf("Wanted %d got %d", http.StatusOK, status)
	}
}

func TestFetchStatusCodeFromPage404(t *testing.T) {
	t.Parallel()

	ts := httptest.NewTLSServer(nil)
	page := ts.URL + "/anything"
	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	lc.HTTPClient = ts.Client()
	status, err := lc.GetPageStatus(page)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusNotFound {
		t.Errorf("Wanted %d got %d", http.StatusNotFound, status)
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
		f, err := os.Open("testdata/links.html")
		if err != nil {
			fmt.Print("error")
		}
		io.Copy(w, f)
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

func TestCheckReturnsAllPages(t *testing.T) {
	ts := httptest.NewTLSServer(http.FileServer(http.Dir("testdata")))
	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	lc.HTTPClient = ts.Client()
	want := []linkchecker.Result{
		{
			Status: http.StatusOK,
			Link:   ts.URL + "/links.html",
		},
		{
			Status: http.StatusOK,
			Link:   ts.URL + "/whatever.html",
		},
		{
			Status: http.StatusNotFound,
			Link:   ts.URL + "/me.html",
		},
		{
			Status: http.StatusOK,
			Link:   ts.URL + "/you.html",
		},
	}
	startLink := ts.URL + "/links.html"
	err = lc.Check(startLink)
	lc.Workers.Wait()
	if err != nil {
		t.Fatal(err)
	}
	got := lc.Links
	if !cmp.Equal(want, got, cmpopts.SortSlices(func(x, y linkchecker.Result) bool {
		return x.Link < y.Link
	})) {
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

func TestIsExternalYes(t *testing.T) {
	t.Parallel()
	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	lc.Domain = "example.com"
	tcs := []string{
		"https://bogus1.com/foo.html",
		"https://bogus2.com/",
		"https://bogus3.com/bar.html",
		"https://bogus.com/search?query=example.com",
		"http://bogus.com/search?query=example.com",
	}
	for _, link := range tcs {
		external := lc.IsExternal(link)
		if !external {
			t.Errorf("not detected as external: %s", link)
		}
	}
}

func TestIsExternalNo(t *testing.T) {
	t.Parallel()
	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	lc.Domain = "example.com"
	tcs := []string{
		"https://example.com/foo.html",
		"https://example.com/",
		"https://example.com/bar.html",
		"http://example.com",
	}
	for _, link := range tcs {
		external := lc.IsExternal(link)
		if external {
			t.Errorf("wrongly detected as external: %s", link)
		}
	}
}

func TestCanonicaliseLinkSameDomain(t *testing.T) {
	t.Parallel()
	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	lc.Domain = "example.com"
	want := "https://example.com/foo.html"
	tcs := []string{
		"foo.html",
		"https://example.com/foo.html",
		"example.com/foo.html",
		// "https://example.com/foo.html?query=example.com",
	}
	for _, input := range tcs {
		got := lc.CanonicaliseLink(input)
		if !cmp.Equal(want, got) {
			t.Error(input, cmp.Diff(want, got))
		}
	}
}

func TestCanonicaliseLinkOtherDomain(t *testing.T) {
	t.Parallel()
	input := "https://bogus.com/"
	want := "https://bogus.com/"
	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	got := lc.CanonicaliseLink(input)
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
