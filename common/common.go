package common

import (
	"crypto/tls"
	"net/http"
	"os"
)

func Fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}

func GetDefaultClient(insecure bool) *http.Client {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: insecure}
	return http.DefaultClient
}
