package testutils

import (
	"net/http"
	"net/http/httptest"
	"os"
)

func NewHttpTestServer(handler http.HandlerFunc) (string, func()) {
	server := httptest.NewServer(handler)
	return server.URL, func() {
		server.Close()
	}
}

func PopulateHttpResponseWithFileContents(w http.ResponseWriter, filename string) error {
	f, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	_, err = w.Write(f)
	if err != nil {
		return err
	}
	return nil
}
