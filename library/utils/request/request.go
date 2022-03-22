// @Author: Vcentor
// @Date: 2020/10/26 4:53 下午
package request

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	FORMConverter = "application/x-www-form-urlencoded"
	JSONConverter = "application/json"
	FILEConverter = "application/octet-stream"
)

type HTTPResp struct {
	Raw         []byte
	ContentType string
}

// HTTPRequester 请求结构体
type HTTPRequester struct {
	Method      string
	Url         string
	ReqParams   []byte
	ContentType string
	Filename    string
}

// NewHTTPRequester 初始化NewHTTPRequester对象
// GET和x-wwww-form-urlencoded的post请求  key-value格式: foo=1&bar=2
// POST pb和json请求：[]byte即可
func NewHTTPRequester(method, url, contentType, filename string, reqParams []byte) *HTTPRequester {
	return &HTTPRequester{
		Method:      method,
		Url:         url,
		ReqParams:   reqParams,
		ContentType: contentType,
		Filename:    filename,
	}
}

// Request 发送请求
func (h *HTTPRequester) Request(httpResp *HTTPResp) error {
	var (
		req = &http.Request{}
		err error
	)
	h.Method = strings.ToUpper(h.Method)
	if h.Method == "GET" || h.Method == "" {
		if req, err = http.NewRequest(h.Method, h.Url, nil); err != nil {
			return err
		}
		m, err := url.ParseQuery(string(h.ReqParams))
		if err != nil {
			return err
		}
		q := req.URL.Query()
		for k, item := range m {
			for _, v := range item {
				q.Add(k, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}
	// key-value
	if h.Method == "POST" && h.ContentType == FORMConverter {
		if req, err = http.NewRequest(h.Method, h.Url, bytes.NewBuffer(h.ReqParams)); err != nil {
			return err
		}
	}
	// key-value 下载文件
	if h.Method == "POST" && h.ContentType == FILEConverter {
		if req, err = http.NewRequest(h.Method, h.Url, bytes.NewBuffer(h.ReqParams)); err != nil {
			return err
		}
		req.Header.Set("Accept-Ranges", "bytes")
		req.Header.Set("Content-Disposition", "attachment; filename="+h.Filename+"")
	}
	// pb or json or multipart/form-data
	if h.Method == "POST" && h.ContentType != FORMConverter && h.ContentType != FILEConverter {
		if req, err = http.NewRequest(h.Method, h.Url, bytes.NewBuffer(h.ReqParams)); err != nil {
			return err
		}
	}
	defer req.Body.Close()
	// 阻止连接被重用，设置http为短连接
	// req.Close = true
	if h.ContentType != "" {
		req.Header.Set("Content-Type", h.ContentType)
	}
	// 设置超时
	client := &http.Client{
		Timeout: time.Duration(60) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		// 重试1次
		time.Sleep(50 * time.Millisecond)
		resp, err = client.Do(req)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()
	statusCode := resp.StatusCode
	if statusCode != 200 {
		return fmt.Errorf("Request [%s] failed!statusCode=%d", h.Url, statusCode)
	}
	b, _ := ioutil.ReadAll(resp.Body)
	httpResp.Raw = b
	httpResp.ContentType = resp.Header.Get("Content-Type")
	return nil
}
