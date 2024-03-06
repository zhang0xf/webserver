package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func TestHttpClient(t *testing.T) {
	resp, err := http.PostForm("http://192.168.5.220:8505/api/whNotice",
		url.Values{"info": {"<span style='font-weight:bold;font:30px' color='#ff0000'>标题</span><br/><span style='font-weight:bold;font:24px' color='#6ad2e3'>此次版本更新内容:XXX</span><br/>"}})

	if err != nil {
		fmt.Println("error : " + err.Error())
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error : " + err.Error())
		return
	}

	fmt.Println(string(body))
}
