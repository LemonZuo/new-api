package openai

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"net/http"
	"net/url"
	"one-api/common"
)

// 官方接口地址
const auth0TokenURL = "https://auth0.openai.com/oauth/token"

// 始皇代理接口地址
const mirrorTokenURL = "https://token.oaifree.com/api/auth/refresh"

// 是否使用代理
const useMirror = true

// RefreshRequest 请求的结构体
type RefreshRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	RefreshToken string `json:"refresh_token"`
	RedirectURI  string `json:"redirect_uri"`
}

// TokenResponse 响应的结构体
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// sendRequest 用于发送HTTP请求并处理响应
func sendRequest(req *http.Request) (TokenResponse, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, err
	}
	defer func() {
		if resp != nil {
			if closeError := resp.Body.Close(); closeError != nil {
				// 在这里处理关闭时的错误
				common.SysError(fmt.Sprintf("关闭响应体失败: %s", closeError))
			}
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenResponse{}, err
	}

	// 打印状态码和响应体
	common.SysLog(fmt.Sprintf("RT -> AT Status Code: %d", resp.StatusCode))
	common.SysLog(fmt.Sprintf("RT -> AT Response Body: %s", string(respBody)))

	if resp.StatusCode != http.StatusOK {
		return TokenResponse{}, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	err = json.Unmarshal(respBody, &tokenResp)
	if err != nil {
		return TokenResponse{}, err
	}

	return tokenResp, nil
}

// refreshTokenUsingAuth0 通过官方接口刷新Token
func refreshTokenUsingAuth0(refreshToken string) (TokenResponse, error) {
	reqData := RefreshRequest{
		RedirectURI:  "com.openai.chat://auth0.openai.com/ios/com.openai.chat/callback",
		GrantType:    "refresh_token",
		ClientID:     "pdlLIX2Y72MIl2rhLhTE9VV9bN905kBh",
		RefreshToken: refreshToken,
	}

	reqBodyBytes, err := json.Marshal(reqData)
	if err != nil {
		return TokenResponse{}, err
	}

	req, err := http.NewRequest("POST", auth0TokenURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return TokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	return sendRequest(req)
}

// refreshTokenUsingMirror 通过镜像接口刷新Token
func refreshTokenUsingMirror(refreshToken string) (TokenResponse, error) {
	formData := url.Values{}
	formData.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", mirrorTokenURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return TokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return sendRequest(req)
}

// RefreshAccessToken 发送请求并解析响应以刷新Access Token
func RefreshAccessToken(refreshToken string) (TokenResponse, error) {
	if len(refreshToken) == 0 {
		return TokenResponse{}, fmt.Errorf("refresh token is empty")
	}

	var res TokenResponse
	var err error
	if useMirror {
		res, err = refreshTokenUsingMirror(refreshToken)
	} else {
		res, err = refreshTokenUsingAuth0(refreshToken)
	}
	if err != nil {
		return TokenResponse{}, err
	}

	// 检查获取的Access Token是否为空
	if len(res.AccessToken) == 0 {
		return TokenResponse{}, fmt.Errorf("get access token failed, the access token is empty")
	}

	return res, nil
}
