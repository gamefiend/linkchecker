package linkchecker_test

import (
	"fmt"
	"io"
	"linkchecker"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFetchStatusCodeFromPage(t *testing.T) {

	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "here is your page")
	}))
	page := s.URL + "/test"
	status, err := linkchecker.GetPageStatus(page, s.Client())
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Errorf("Wanted %d got %d", http.StatusOK, status)
	}
}

func TestGrabLinksFromPage(t *testing.T) {
	want := []string{"whatever", "you"}
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
	want := []string{"whatever.html", "you.html"}
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, err := os.ReadFile("testdata/links.html")
		if err != nil {
			fmt.Print("error")
		}
		io.Copy(w, strings.NewReader(string(file)))
	}))
	got, err := linkchecker.GrabLinksFromServer(s.URL, s.Client())
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Errorf("want %v, got: %v", want, got)
	}
}

func TestCheckLinksReturnsAllPages(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveFile := "testdata/links.html"
		if strings.Contains(r.RequestURI, "whatever.html") {
			serveFile = "testdata/whatever.html"
		}
		if strings.Contains(r.RequestURI, "me.html") {
			w.WriteHeader(404)
		}
		if strings.Contains(r.RequestURI, "you.html") {
			serveFile = "testdata/you.html"
		}
		file, err := os.ReadFile(serveFile)
		if err != nil {
			fmt.Print("error")
		}
		io.Copy(w, strings.NewReader(string(file)))
	}))
	linkchecker.Domain = s.URL
	want := []linkchecker.Link{
		{
			Status: 200,
			URL:    s.URL,
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

	err := linkchecker.CheckLinks(s.URL, s.Client())
	if err != nil {
		t.Fatal(err)
	}
	got := linkchecker.Links
	if !cmp.Equal(want, got) {
		t.Errorf("want %v, got: %v", want, got)
	}
}
