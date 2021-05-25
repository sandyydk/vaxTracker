package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

var client *http.Client
var once sync.Once

const (
	maxConnPerHost = 10
	maxIdleConns   = 5
)

func initializeClient() {
	timeout := time.Duration(60 * time.Second)
	client = &http.Client{
		Timeout: timeout,
	}

	// Disable the default SSL check
	transport := http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		MaxConnsPerHost:   maxConnPerHost,
		MaxIdleConns:      maxIdleConns,
	}

	client.Transport = &transport

}

// GetHTTPRequest returns a request object
func GetHTTPRequest(method, endpoint string, body interface{}) (*http.Request, error) {

	var writer io.ReadWriter

	if body != nil {
		writer = new(bytes.Buffer)
		err := json.NewEncoder(writer).Encode(body)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	req, err := http.NewRequest(method, endpoint, writer)
	if err != nil {
		log.Println("Error:", err)
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Add("User-Agent", "vaxTracker")

	return req, nil
}

// ExecuteHTTPRequest executes a request to the target
func ExecuteHTTPRequest(req *http.Request) (*http.Response, error) {

	once.Do(initializeClient)
	defer client.CloseIdleConnections()

	response, err := client.Do(req)

	return response, err
}
