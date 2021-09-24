package linkchecker_test

import (
	"fmt"
	"linkchecker"
	"net/http"
	"net/http/httptest"
	"testing"
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
