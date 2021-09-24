package linkchecker

import (
	"net/http"
)

func GetPageStatus(page string, client *http.Client) (int, error) {
	resp, err := client.Get(page)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}
