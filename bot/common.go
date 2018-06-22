package bot

import (
	"fmt"
	"net/http"
	"strings"
)

func parse1Params(r *http.Request, prefix string, num int) ([]string, error) {
	url := strings.TrimPrefix(r.URL.Path, prefix)
	params := strings.Split(url, "/")
	if len(params) != num || len(params[0]) == 0 {
		return nil, fmt.Errorf("Bad format. Expecting exactly %d params", num)
	}
	return params, nil
}

func parse2Params(r *http.Request, prefix string, num int) ([]string, error) {
	url := strings.TrimPrefix(r.URL.Path, prefix)
	params := strings.Split(url, "/")
	if len(params) != num || len(params[0]) == 0 || len(params[1]) == 0 {
		return nil, fmt.Errorf("Bad format. Expecting exactly %d params", num)
	}
	return params, nil
}

func parse3Params(r *http.Request, prefix string, num int) ([]string, error) {
	url := strings.TrimPrefix(r.URL.Path, prefix)
	params := strings.Split(url, "/")
	if len(params) != num || len(params[0]) == 0 || len(params[1]) == 0 || len(params[2]) == 0 {
		return nil, fmt.Errorf("Bad format. Expecting exactly %d params", num)
	}
	return params, nil
}
