package linkchecker_test

import (
	"encoding/json"
	"linkchecker"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFormatTermProvidesCorrectOutput(t *testing.T) {
	t.Parallel()

	// this is a lot of bootstrapping to avoid having t do weird formatting on the want value, will reconsider this
	ts := httptest.NewTLSServer(http.FileServer(http.Dir("testdata")))

	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	lc.HTTPClient = ts.Client()
	startLink := ts.URL + "/links.html"
	err = lc.Check(startLink)
	lc.Workers.Wait()
	if err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder

	sb.WriteString("200 " + ts.URL + "/links.html\n")
	sb.WriteString("404 " + ts.URL + "/me.html\n")
	sb.WriteString("200 " + ts.URL + "/whatever.html\n")
	sb.WriteString("200 " + ts.URL + "/you.html\n")

	want := sb.String()
	var lt linkchecker.LinksTerminal
	got, err := lt.Format(lc)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestFormatJSONProvidesCorrectOutput(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(http.FileServer(http.Dir("testdata")))
	lc, err := linkchecker.New()
	if err != nil {
		t.Fatal(err)
	}
	startLink := ts.URL + "/links.html"
	lc.HTTPClient = ts.Client()
	err = lc.Check(startLink)
	lc.Workers.Wait()
	if err != nil {
		t.Fatal(err)
	}
	wl := []linkchecker.Result{
		{
			Status: 200,
			Link:   ts.URL + "/links.html",
		},
		{
			Status: 200,
			Link:   ts.URL + "/whatever.html",
		},
		{
			Status: 404,
			Link:   ts.URL + "/me.html",
		},
		{
			Status: 200,
			Link:   ts.URL + "/you.html",
		},
	}
	sort.Slice(wl, func(i, j int) bool {
		return wl[i].Link < wl[j].Link
	})
	want, _ := json.Marshal(wl)
	var lj linkchecker.LinksJSON

	got, err := lj.Format(lc)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(string(want), got) {
		t.Error(cmp.Diff(string(want), got))
	}
}
