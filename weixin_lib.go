package ginlib

import (
	"bytes"
	"crypto"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"time"
)

//WxGetOpenId 微信获取openid
func WxGetOpenId(code string) (openid, accessToken string) {
	appId := Ini_Str("weixin.app_id")
	secret := Ini_Str("weixin.app_secret")

	resp, err := http.Get(fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", appId, secret, code))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic(fmt.Errorf("微信获取openid失败, statusCode:%d", resp.StatusCode))
	}
	//读取结果
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	//解析结果
	result := gin.H{}
	if err := json.Unmarshal(res, &result); err != nil {
		panic(err)
	}
	openid = result["openid"].(string)
	accessToken = result["access_token"].(string)

	return
}

// WxGetUserInfo 获取微信用户信息
func WxGetUserInfo(openid, accessToken string) (result gin.H) {
	resp, err := http.Get(fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s", accessToken, openid))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic(fmt.Errorf("获取微信用户信息失败, statusCode:%d", resp.StatusCode))
	}
	//读取结果
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	//解析结果
	if err := json.Unmarshal(res, &result); err != nil {
		panic(err)
	}
	return
}

//todo WxPhoneCodeSend 腾讯发送验证码
func WxPhoneCodeSend(phone, code string) (err error) {
	return
}

// WxPaymentOrder 微信支付
func WxPaymentOrder(orderNo string, amount int, productDesc, notifyUrl string) (prepayId string) {
	data := gin.H{
		"appid":Ini_Str("weixin.app_id"),
		"mchid":Ini_Str("weixin.mchid"),
		"description":productDesc,
		"out_trade_no":orderNo,
		"notify_url":notifyUrl,
		"amount":gin.H{
			"total":amount,
		},
	}
	datajson, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	//发起请求
	resp, err := http.Post("https://api.mch.weixin.qq.com/v3/pay/transactions/app", "application/json", bytes.NewReader(datajson))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic(fmt.Errorf("发起微信支付失败, statusCode:%d", resp.StatusCode))
	}
	//读取结果
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	//解析结果
	result := gin.H{}
	if err := json.Unmarshal(res, &result); err != nil {
		panic(err)
	}

	prepayId = result["prepay_id"].(string)
	return 
}

//WxPaymentParam 构造客户端唤起参数
func WxPaymentParam(prepayId string) (result gin.H) {
	appId := Ini_Str("weixin.app_id")
	mchId := Ini_Str("weixin.mchid")
	noncestr := UniqueId()
	timestamp := time.Now().Second()

	sign := RsaSign(fmt.Sprintf("%s\n%d\n%s\n%s\n", appId, timestamp, noncestr, prepayId), Ini_Str("weixin.mch_pem_key"), crypto.SHA256)

	return gin.H{
		"appid":appId,
		"partnerid":mchId,
		"prepayid":prepayId,
		"package":"Sign=WXPay",
		"noncestr":noncestr,
		"timestamp":timestamp,
		"sign":sign,
	}
}

//WxPaymentResult 微信支付结果查询
func WxPaymentResult(orderNo string) (tradeState string) {
	mchId := Ini_Str("weixin.mchid")
	//发起查询
	resp, err := http.Get(fmt.Sprintf("https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/%s?mchid=%s", orderNo, mchId))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic(fmt.Errorf("微信支付结果查询失败, statusCode:%d", resp.StatusCode))
	}
	//读取结果
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	//解析结果
	result := gin.H{}
	if err := json.Unmarshal(res, &result); err != nil {
		panic(err)
	}

	return result["trade_state"].(string)
}