package utils

import (
	"net/http"
	"strings"
	"testing"
)

func TestGetNextProxyHeader(t *testing.T) {

	testCases := []map[string][]string{
		//  Remove the first stop
		map[string][]string{
			"input":  []string{"a.com:19090;b.com:19090;c.com:19090", "/query"},
			"expect": []string{"a.com:19090", "b.com:19090;c.com:19090", "/query"},
		},
		// Send to the last stop server
		map[string][]string{
			"input":  []string{"a.com:19090", "/query"},
			"expect": []string{"a.com:19090", "", "/query"},
		},
		// Direct send request to local prometheus server
		map[string][]string{
			"input":  []string{"", "/query"},
			"expect": []string{"", "", "/query"},
		},
		map[string][]string{
			"input":  []string{"", ""},
			"expect": []string{"", "", ""},
			"error":  []string{"return error"},
		},
	}
	for _, tc := range testCases {
		req, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader("foo=bar"))
		req.Header.Set("X-Prometheus-Proxy", tc["input"][0])
		req.Header.Set("X-Prometheus-Request", tc["input"][1])

		parseResult, err := GetNextProxyHeader(req)

		if tc["input"][1] == "" {
			if err == nil {
				t.Fatalf("Expect error when query empty but got nothing %v", parseResult)
			}
			continue
		}
		if len(parseResult) != 3 {
			t.Fatalf("Result lens is %d, expect 3 %v : %v ", len(parseResult), tc["expect"], parseResult)
		}
		if parseResult[0] != tc["expect"][0] || parseResult[1] != tc["expect"][1] || parseResult[2] != tc["expect"][2] {
			t.Fatalf("Expect %v, but got %v", tc["expect"], parseResult)
		}
	}

}
