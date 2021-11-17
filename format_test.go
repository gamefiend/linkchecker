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
	s := httptest.NewTLSServer(http.FileServer(http.Dir("testdata")))

	lc, err := linkchecker.New(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	startLink := s.URL + "/links.html"
	err = lc.CheckLinks(startLink, s.Client())
	lc.Workers.Wait()
	if err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder

	sb.WriteString("200 " + s.URL + "/links.html\n")
	sb.WriteString("404 " + s.URL + "/me.html\n")
	sb.WriteString("200 " + s.URL + "/whatever.html\n")
	sb.WriteString("200 " + s.URL + "/you.html\n")

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
	s := httptest.NewTLSServer(http.FileServer(http.Dir("testdata")))

	lc, err := linkchecker.New(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	startLink := s.URL + "/links.html"
	err = lc.CheckLinks(startLink, s.Client())
	lc.Workers.Wait()
	if err != nil {
		t.Fatal(err)
	}
	wl := []linkchecker.Link{
		{
			Status: 200,
			URL:    s.URL + "/links.html",
		},
		{
			Status: 200,
			URL:    s.URL + "/whatever.html",
		},
		{
			Status: 404,
			URL:    s.URL + "/me.html",
		},
		{
			Status: 200,
			URL:    s.URL + "/you.html",
		},
	}
	sort.Slice(wl, func(i, j int) bool {
		return wl[i].URL < wl[j].URL
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
