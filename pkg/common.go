package pkg

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func Fatal(doing string, err error) {
	fmt.Fprintf(os.Stderr, "error %s: %s\n", doing, err)
	os.Exit(1)
}

func GetDefaultClient(insecure bool) *http.Client {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: insecure}
	return http.DefaultClient
}

func matchPath(patterns []string, path string) bool {
	for _, pattern := range patterns {
		ok, _ := filepath.Match(pattern, path)
		if ok {
			return true
		}
	}
	return false
}
