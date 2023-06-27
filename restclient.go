package gorestlib

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/druvainc/gorestlib/restliberror"
)

// RestClient to Call HTTP endpoints of products
type RestClient struct {
	apiRootURL string
	client     *http.Client
}

// transportConfig ...
func transportConfig() *http.Transport {
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 30 * time.Second, //TCP Connect Timeout 30 Secs
		}).DialContext,
		MaxIdleConnsPerHost: 8,
		MaxIdleConns:        16,
	}
	return tr
}

// SetHeaders ...
func SetHeaders(req *http.Request, headers map[string]string) {
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

// NewRestClient create new instance of RestClient
func NewRestClient(apiRootURL string) *RestClient {
	return &RestClient{
		apiRootURL: apiRootURL,
		client: &http.Client{
			Transport: transportConfig(),
			Timeout:   600 * time.Second}, //Total Read Timeout 10 mins
	}
}

// NewLogger returns new logger with UTC timezone and package prefix
func NewLogger() *log.Logger {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC)
	logger.SetPrefix("gorestlibpkg ")
	return logger
}

// RestClientInterface ...
type RestClientInterface interface {
	Get(resource string,
		responseObj interface{},
		queryParam map[string]string,
		headers map[string]string) error
	Post(resource string,
		bodyParam interface{},
		responseObj interface{},
		headers map[string]string) error
}

// Get for making http GET request
func (r *RestClient) Get(resource string, responseObj interface{}, queryParam map[string]string, headers map[string]string) error {
	logger := NewLogger()
	var err error
	var resp *http.Response
	endpoint := r.apiRootURL + resource
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		logger.Printf("Get: Error in http.NewRequest: Endpoint:%s, Error:%v", endpoint, err)
		return err
	}

	q := req.URL.Query()
	for k, v := range queryParam {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	if headers != nil {
		SetHeaders(req, headers)
	}

	logger.Printf("Get: Debug Log, URL:%v, FilterParams:%v", req.URL, queryParam)

	return r.ProcessResponse(req, resp, responseObj)
}

// Post for making http POST request
func (r *RestClient) Post(resource string, bodyParam interface{}, responseObj interface{}, headers map[string]string) error {
	logger := NewLogger()
	var err error
	var resp *http.Response
	var inputData []byte

	endpoint := r.apiRootURL + resource

	inputData, err = json.Marshal(bodyParam)
	if err != nil {
		logger.Printf("Post: Error in json.Marshal, Endpoint: %s, bodyParam: %v, Error: %v", endpoint, bodyParam, err)
		return err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(inputData))
	if err != nil {
		logger.Printf("Post: Error in http.NewRequest, Endpoint: %s, bodyParam: %v, Error: %v", endpoint, bodyParam, err)
		return err
	}

	if headers != nil {
		SetHeaders(req, headers)
	}

	logger.Printf("Post: Debug Log, URL:%v, bodyParam:%v", req.URL, bodyParam)
	return r.ProcessResponse(req, resp, responseObj)
}

// ProcessResponse to process response of all http requests
func (r *RestClient) ProcessResponse(req *http.Request, resp *http.Response, responseObj interface{}) error {
	logger := NewLogger()
	var err error
	var body []byte

	req.Close = true

	resp, err = r.client.Do(req)
	if err != nil {
		logger.Printf("ProcessResponse: Error in r.client.Do, Method:%v, URL:%v, Error:%v", req.Method, req.URL, err)
		return err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("ProcessResponse: Error in ioutil.ReadAll, Method:%v, URL:%v, Error:%v", req.Method, req.URL, err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
			restLibError := restliberror.RestLibError{
				Err:  errors.New(resp.Status),
				Code: int64(resp.StatusCode),
			}
			return restLibError
		}
		logger.Printf("ProcessResponse: Error returned by http request: Method:%v, URL:%v, Status:%s, ResponseBody: %s", req.Method, req.URL, resp.Status, string(body))
		return errors.New(string(body))
	}

	if len(body) == 0 {
		logger.Printf("ProcessResponse: Recieved empty response body, Method:%v, URL:%v", req.Method, req.URL)
		return nil
	}

	err = json.Unmarshal(body, responseObj)
	if err != nil {
		logger.Printf("ProcessResponse: Error in json.Unmarshal, Method:%v, URL:%v, Error:%v", req.Method, req.URL, err)
		return err
	}

	return nil
}
