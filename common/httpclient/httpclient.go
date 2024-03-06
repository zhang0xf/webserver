package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/*
http://stackoparamserflow.com/questions/16895294/how-to-set-timeout-for-http-get-requests-in-golang
http://stackoverflow.com/questions/12756782/go-http-post-and-use-cookies
*/

var DnsCacheDuration time.Duration = 0
var dnsCache = &DnsCache{caches: make(map[string]DnsCacheItem)}
var DefaultClient *http.Client = NewTimeoutClient(10*time.Second, 10*time.Second)

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		if DnsCacheDuration > 0 {
			addr = dnsCache.Get(addr)
		}
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		return NewTimeoutConn(conn, rwTimeout), nil
	}
}

func NewTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial:  TimeoutDialer(connectTimeout, readWriteTimeout),
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

// 包对外提供的接口
func DoRequest(urlStr string) ([]byte, error) {
	resp, err := DefaultClient.Get(urlStr)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, err
}

func DoGet(urlStr string, params url.Values) ([]byte, error) {
	if params != nil {
		if strings.Contains(urlStr, "?") {
			urlStr += "&" + params.Encode()
		} else {
			urlStr += "?" + params.Encode()
		}
	}
	fmt.Println("http DoGet:", urlStr)

	resp, err := DefaultClient.Get(urlStr)
	if err != nil {
		fmt.Println("DoGet DefaultClient.Get error", err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println("DoGet ioutil.ReadAll error", err)
		return nil, err
	}
	return body, err
}

func DoGetEx(urlStr string, params url.Values, p interface{}) error {
	rb, err := DoGet(urlStr, params)
	if err != nil {
		fmt.Println("DoGetEx get error", err)
		return err
	}
	err = json.Unmarshal(rb, p)
	if err != nil {
		fmt.Println("DoGetEx Unmarshal error", rb, err)
	}
	return err
}

func DoPost(urlStr string, params url.Values) ([]byte, error) {
	var postReader io.Reader = nil
	if params != nil {
		postReader = strings.NewReader(params.Encode())
	}
	resp, err := DefaultClient.Post(urlStr, "application/x-www-form-urlencoded", postReader)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, err
}

func DoPostEx(urlStr string, params url.Values, p interface{}) error {
	rb, err := DoPost(urlStr, params)
	if err != nil {
		return err
	}
	err = json.Unmarshal(rb, p)
	return err
}

func DoJsonPost(urlStr string, param interface{}) ([]byte, error) {
	requestJson, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	resp, err := DefaultClient.Post(urlStr, "application/json; charset=UTF-8", bytes.NewBuffer(requestJson))
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, err
}

func DoMultiPartPost(urlStr string, params url.Values, files url.Values) ([]byte, error) {
	var err error
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, values := range files {
		fileName := values[0]
		if strings.HasPrefix(fileName, "http://") {
			resp, err := DefaultClient.Get(fileName)
			if err != nil {
				return nil, err
			}
			part, err := writer.CreateFormFile(key, "pic")
			if err != nil {
				return nil, err
			}
			// _, err = io.Copy(part, resp.Body)
			io.Copy(part, resp.Body)
			defer resp.Body.Close()
		} else {
			file, err := os.Open(fileName)
			if err != nil {
				return nil, err
			}
			defer file.Close()
			part, err := writer.CreateFormFile(key, filepath.Base(fileName))
			if err != nil {
				return nil, err
			}
			// _, err = io.Copy(part, file)
			io.Copy(part, file)
		}
	}
	for key, values := range params {
		_ = writer.WriteField(key, values[0])
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body2, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body2, nil
}

func DoMultiPartPostEx(urlStr string, params url.Values, files url.Values, p interface{}) error {
	rb, err := DoMultiPartPost(urlStr, params, files)
	if err != nil {
		return err
	}
	err = json.Unmarshal(rb, p)
	return err
}

func Do(req *http.Request) (*http.Response, error) {
	return DefaultClient.Do(req)
}
