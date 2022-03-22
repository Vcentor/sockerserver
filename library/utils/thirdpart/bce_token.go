// @Author: Vcentor
// @Date: 2020/12/16 8:54 下午

package thirdpart

import (
	"socketserver/library/utils/request"
	"encoding/json"
	"fmt"
)

const (
	TOKEN_URL = "https://aip.baidubce.com/oauth/2.0/token"
)

// AccessToken bce服务返回数据
type AccessToken struct {
	RefreshToken  string `json:"refresh_token"`
	ExpiresIn     int    `json:"expires_in"`
	Scope         string `json:"scope"`
	SessionKey    string `json:"session_key"`
	Token         string `json:"access_token"`
	SessionSecret string `json:"session_secret"`
}

// BceAccessToken 获取百度AI平台access_token
func BceAccessToken(appKey, secretKey string) (string, error) {
	var token string
	params := fmt.Sprintf("grant_type=%s&client_id=%s&client_secret=%s", "client_credentials", appKey, secretKey)
	var tokenResp request.HTTPResp
	if err := request.NewHTTPRequester("POST", TOKEN_URL, request.FORMConverter, "", []byte(params)).Request(&tokenResp); err != nil {
		return token, fmt.Errorf("get access_token request error %s", err.Error())
	}
	var accessToken AccessToken
	if err := json.Unmarshal(tokenResp.Raw, &accessToken); err != nil {
		return token, fmt.Errorf("get access_token unmarshal response error %s", err.Error())
	}
	token = accessToken.Token
	return token, nil
}
