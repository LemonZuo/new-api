package channel

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"io"
	"net/http"
	logCommon "one-api/common"
	"one-api/relay/common"
	"one-api/service"
)

func SetupApiRequestHeader(info *common.RelayInfo, c *gin.Context, req *http.Request) {
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))
	if info.IsStream && c.Request.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "text/event-stream")
	}
	// set customer headers
	if len(info.Headers) <= 0 {
		return
	}

	// json unmarshal headers
	headers := make(map[string]string)
	err := json.Unmarshal([]byte(info.Headers), &headers)
	if err != nil {
		logCommon.LogError(c, "unmarshal_headers_failed")
		return
	}

	// loop through the map and set the headers
	for k, v := range headers {
		if len(v) > 0 {
			req.Header.Set(k, v)
		}
	}
}

func DoApiRequest(a Adaptor, c *gin.Context, info *common.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	fullRequestURL, err := a.GetRequestURL(info)
	if err != nil {
		return nil, fmt.Errorf("get request url failed: %w", err)
	}
	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("new request failed: %w", err)
	}
	err = a.SetupRequestHeader(c, req, info)
	if err != nil {
		return nil, fmt.Errorf("setup request header failed: %w", err)
	}

	var resp *http.Response

	if len(info.Proxy) <= 0 {
		resp, err = doRequest(c, req)
	} else {
		resp, err = doRequestWithProxy(c, req, info.Proxy)
	}
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	return resp, nil
}

func doRequest(c *gin.Context, req *http.Request) (*http.Response, error) {
	resp, err := service.GetHttpClient().Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("resp is nil")
	}
	_ = req.Body.Close()
	_ = c.Request.Body.Close()
	return resp, nil
}

func doRequestWithProxy(c *gin.Context, req *http.Request, proxy string) (*http.Response, error) {
	client, err := service.GetProxyHttpClient(proxy)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("resp is nil")
	}
	_ = req.Body.Close()
	_ = c.Request.Body.Close()
	return resp, nil
}
