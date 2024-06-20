package service

import (
	"context"
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
	"one-api/common"
	"strings"
	"time"
)

var httpClient *http.Client
var impatientHTTPClient *http.Client

func init() {
	if common.RelayTimeout == 0 {
		httpClient = &http.Client{}
	} else {
		httpClient = &http.Client{
			Timeout: time.Duration(common.RelayTimeout) * time.Second,
		}
	}

	impatientHTTPClient = &http.Client{
		Timeout: 5 * time.Second,
	}
}

func GetHttpClient() *http.Client {
	return httpClient
}

func GetImpatientHttpClient() *http.Client {
	return impatientHTTPClient
}

func GetProxyHttpClient(proxyURLStr string) (*http.Client, error) {
	// 解析代理URL
	proxyURL, err := url.Parse(proxyURLStr)
	if err != nil {
		return nil, fmt.Errorf("解析代理URL失败: %v", err)
	}

	// 获取代理的认证信息（如果有）
	auth := &proxy.Auth{}
	if proxyURL.User != nil {
		auth.User = proxyURL.User.Username()
		password, isSet := proxyURL.User.Password()
		if isSet {
			auth.Password = password
		}
	}

	// 检查代理协议是否为socks5
	if strings.HasPrefix(proxyURL.Scheme, "socks5") {
		// 使用认证信息创建SOCKS5代理
		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}

		// 创建Transport
		transport := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
		}

		// 创建并返回配置了SOCKS5代理的http.Client
		return &http.Client{Transport: transport}, nil
	} else {
		// 对于HTTP代理，需要设置代理的HTTP头
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

		// 创建并返回配置了HTTP代理的http.Client
		return &http.Client{
			Transport: transport,
		}, nil
	}
}
