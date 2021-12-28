package ginlib

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

//WxGetOpenId 微信获取openid
func WxGetOpenId(code string) (openid, accessToken string, err error) {
	appId := Ini_Str("weixin.app_id")
	secret := Ini_Str("weixin.app_secret")

	if appId == "" || secret == "" {
		err = fmt.Errorf("请配置必须参数, [weixin.app_id], [weixin.app_secret]")
		return
	}

	resp, err := http.Get(fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", appId, secret, code))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("微信获取openid失败, statusCode:%d", resp.StatusCode)
		return
	}
	//读取结果
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	//解析结果
	result := gin.H{}
	if err = json.Unmarshal(res, &result); err != nil {
		return
	}
	openid = result["openid"].(string)
	accessToken = result["access_token"].(string)

	return
}

type WxGetUserInfoResponse struct {
	OpenId     string `json:"openid"`
	NickName   string `json:"nickname"`
	Sex        int32  `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadImgUrl string `json:"headimgurl"`
	UnionId    string `json:"unionid"`
}

// WxGetUserInfo 获取微信用户信息
func WxGetUserInfo(openid, accessToken string) (result WxGetUserInfoResponse, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s", accessToken, openid))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("获取微信用户信息失败, statusCode:%d", resp.StatusCode)
		return
	}
	//读取结果
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	//解析结果
	if err = json.Unmarshal(res, &result); err != nil {
		return
	}
	return
}