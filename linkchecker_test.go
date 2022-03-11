package linkchecker_test

import (
	"encoding/json"
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
	status, err := lc.GetHTTPStatus(page)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Errorf("Wanted %d got %d", http.StatusOK, status)
	}
}

func TestDetects404Status(t *testing.T) {
	t.Parallel()

	ts := httptest.NewTLSServer(nil)
	page := ts.URL + "/anything"
	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	lc.HTTPClient = ts.Client()
	status, err := lc.GetHTTPStatus(page)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusNotFound {
		t.Errorf("Wanted %d got %d", http.StatusNotFound, status)
	}
}

func TestGrabsCorrectLinksFromPage(t *testing.T) {
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
	lc, err := linkchecker.New(
		linkchecker.WithJSONOutput())
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

func TestCheckReturnsAllPagesStreamingDefault(t *testing.T) {
	ts := httptest.NewTLSServer(http.FileServer(http.Dir("testdata")))
	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	lc.HTTPClient = ts.Client()
	want := []linkchecker.Result{
		{
			LinkStatus: linkchecker.LinkStatusCritical,
			HTTPStatus: http.StatusNotFound,
			Link:       ts.URL + "/me.html",
		},
	}
	startLink := ts.URL + "/links.html"
	err = lc.Check(startLink)

	if err != nil {
		t.Fatal(err)
	}
	var got []linkchecker.Result
	var stream <-chan linkchecker.Result
	stream = lc.StreamResults()
	for result := range stream {
		got = append(got, result)
	}

	if !cmp.Equal(want, got, cmpopts.SortSlices(func(x, y linkchecker.Result) bool {
		return x.Link < y.Link
	})) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestCheckReturnsAllPagesStreamingVerbose(t *testing.T) {
	ts := httptest.NewTLSServer(http.FileServer(http.Dir("testdata")))
	lc, err := linkchecker.New(linkchecker.WithVerboseOutput())
	if err != nil {
		t.Fatal(err)
	}
	lc.HTTPClient = ts.Client()
	want := []linkchecker.Result{
		{
			HTTPStatus: http.StatusOK,
			LinkStatus: linkchecker.LinkStatusOK,
			Link:       ts.URL + "/links.html",
		},
		{
			HTTPStatus: http.StatusOK,
			LinkStatus: linkchecker.LinkStatusOK,
			Link:       ts.URL + "/whatever.html",
		},
		{
			HTTPStatus: http.StatusNotFound,
			LinkStatus: linkchecker.LinkStatusCritical,
			Link:       ts.URL + "/me.html",
		},
		{
			HTTPStatus: http.StatusOK,
			LinkStatus: linkchecker.LinkStatusOK,
			Link:       ts.URL + "/you.html",
		},
	}
	startLink := ts.URL + "/links.html"
	err = lc.Check(startLink)

	if err != nil {
		t.Fatal(err)
	}
	var got []linkchecker.Result
	var stream <-chan linkchecker.Result
	stream = lc.StreamResults()
	for result := range stream {
		got = append(got, result)
	}

	if !cmp.Equal(want, got, cmpopts.SortSlices(func(x, y linkchecker.Result) bool {
		return x.Link < y.Link
	})) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestCheckReturnsAllPagesAllResults(t *testing.T) {
	ts := httptest.NewTLSServer(http.FileServer(http.Dir("testdata")))
	lc, err := linkchecker.New(linkchecker.WithVerboseOutput())
	if err != nil {
		t.Fatal(err)
	}
	lc.HTTPClient = ts.Client()
	want := []linkchecker.Result{
		{
			HTTPStatus: http.StatusOK,
			LinkStatus: linkchecker.LinkStatusOK,
			Link:       ts.URL + "/links.html",
		},
		{
			HTTPStatus: http.StatusOK,
			LinkStatus: linkchecker.LinkStatusOK,
			Link:       ts.URL + "/whatever.html",
		},
		{
			HTTPStatus: http.StatusNotFound,
			LinkStatus: linkchecker.LinkStatusCritical,
			Link:       ts.URL + "/me.html",
		},
		{
			HTTPStatus: http.StatusOK,
			LinkStatus: linkchecker.LinkStatusOK,
			Link:       ts.URL + "/you.html",
		},
	}
	startLink := ts.URL + "/links.html"
	err = lc.Check(startLink)

	if err != nil {
		t.Fatal(err)
	}
	var got []linkchecker.Result
	got = lc.AllResults()

	if !cmp.Equal(want, got, cmpopts.SortSlices(func(x, y linkchecker.Result) bool {
		return x.Link < y.Link
	})) {
		t.Error(cmp.Diff(want, got))
	}
}

// func TestUnparseableURLIsReported(t *testing.T) {
// 	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintln(w, `<a href="bogus:// unparseable">bogus</a>`)
// 	}))
// 	lc, err := linkchecker.New(linkchecker.WithVerboseOutput())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	lc.HTTPClient = ts.Client()
// 	want := []linkchecker.Result{
// 		{
// 			Status:     linkchecker.StatusOK,
// 			HTTPStatus: http.StatusOK,
// 			Link:       ts.URL,
// 		},
// 		{
// 			Status: linkchecker.StatusCritical,
// 			Link:   ts.URL + "/bogus://unparseable",
// 		},
// 	}
// 	err = lc.Check(ts.URL)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	var got []linkchecker.Result
// 	got = lc.AllResults()
// 	slices.SortFunc(got, func(x, y linkchecker.Result) bool {
// 		return x.Link < y.Link
// 	})
// 	if !cmp.Equal(want, got) {
// 		t.Error(cmp.Diff(want, got))
// 	}
// }

func TestDetectsLinksOutsideDomain(t *testing.T) {
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

func TestDetectsLinksInsideDomain(t *testing.T) {
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

// func TestCanonicaliseLinkUnparseable(t *testing.T) {
// 	t.Parallel()
// 	lc, err := linkchecker.New()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	want := "bogus://unparseable"
// 	got := lc.CanonicaliseLink(want)
// 	if !cmp.Equal(want, got) {
// 		t.Error(want, cmp.Diff(want, got))
// 	}
// }

func TestResultsPrintAStringNicely(t *testing.T) {
	t.Parallel()
	r := linkchecker.Result{
		LinkStatus: linkchecker.LinkStatusOK,
		Link:       "https://example.com",
	}
	want := "https://example.com OK"
	got := r.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestResultsCreateAProperJSONObject(t *testing.T) {
	t.Parallel()
	want := linkchecker.Result{
		Link:       "https://example.com",
		LinkStatus: linkchecker.LinkStatusOK,
		HTTPStatus: http.StatusOK,
	}
	j := want.ToJSON()
	got := linkchecker.Result{}
	err := json.Unmarshal([]byte(j), &got)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
func TestMapHTTPToLinkReturnsProperResults(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name  string
		input int
		want  linkchecker.LinkStatus
	}{
		{
			name:  "StatusOK -> LinkStatusOK",
			input: http.StatusOK,
			want:  linkchecker.LinkStatusOK,
		},
		{
			name:  "StatusNotFound -> LinkStatusCritical",
			input: http.StatusNotFound,
			want:  linkchecker.LinkStatusCritical,
		},
		{
			name:  "StatusInternalServerErr -> LinkStatusWarning",
			input: http.StatusInternalServerError,
			want:  linkchecker.LinkStatusWarning,
		},
	}

	for _, tc := range tcs {
		got := linkchecker.MapHTTPToLink(tc.input)
		if tc.want != got {
			t.Errorf("%s\nWanted: %v	Got: %v", tc.name, tc.want, got)
		}
	}
}
