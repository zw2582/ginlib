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
func WxGetOpenId(code string) (openid, accessToken string, err error) {
	appId := Ini_Str("weixin.app_id")
	secret := Ini_Str("weixin.app_secret")

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

//todo WxPhoneCodeSend 腾讯发送验证码
func WxPhoneCodeSend(phone, code string) (err error) {
	return
}

// WxPaymentOrder 微信支付
// @link https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_2_1.shtml
func WxPaymentOrder(orderNo string, amount int, productDesc, notifyUrl string) (prepayId string, err error) {
	data := gin.H{
		"appid":        Ini_Str("weixin.app_id"),
		"mchid":        Ini_Str("weixin.mchid"),
		"description":  productDesc,
		"out_trade_no": orderNo,
		"notify_url":   notifyUrl,
		"amount": gin.H{
			"total": amount,
		},
	}
	datajson, err := json.Marshal(data)
	if err != nil {
		return
	}
	//发起请求
	resp, err := http.Post("https://api.mch.weixin.qq.com/v3/pay/transactions/app", "application/json", bytes.NewReader(datajson))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("发起微信支付失败, statusCode:%d", resp.StatusCode)
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

	prepayId = result["prepay_id"].(string)
	return
}

//WxPaymentParam 构造客户端唤起参数
// @link https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_2_1.shtml
func WxPaymentParam(prepayId string) (result gin.H) {
	appId := Ini_Str("weixin.app_id")
	mchId := Ini_Str("weixin.mchid")
	noncestr := UniqueId()
	timestamp := time.Now().Second()

	sign := RsaSign(fmt.Sprintf("%s\n%d\n%s\n%s\n", appId, timestamp, noncestr, prepayId), Ini_Str("weixin.mch_pem_key"), crypto.SHA256)

	return gin.H{
		"appid":     appId,
		"partnerid": mchId,
		"prepayid":  prepayId,
		"package":   "Sign=WXPay",
		"noncestr":  noncestr,
		"timestamp": timestamp,
		"sign":      sign,
	}
}

type WxPaymentResult struct {
	Amount struct{
		Currency string `json:"currency"`
		PayerCurrency string `json:"payer_currency"`
		PayerTotal int `json:"payer_total"`
		Total int `json:"total"`
	} `json:"amount"`
	AppId string `json:"appid"`
	Attach string `json:"attach"`
	BankType string `json:"bank_type"`
	Mchid string `json:"mchid"`
	OutTradeNo string `json:"out_trade_no"`
	Payer struct{
		OpenId string `json:"openid"`
	} `json:"payer"`
	SuccessTime string `json:"success_time"`
	TradeState string `json:"trade_state"`
	TradeStateDesc string `json:"trade_state_desc"`
	TradeType string `json:"trade_type"`
	TransactionId string `json:"transaction_id"`
}

//WxPaymentResultGet 微信支付结果查询
// @link https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_2_1.shtml
func WxPaymentResultGet(orderNo string) (result WxPaymentResult, err error) {
	mchId := Ini_Str("weixin.mchid")
	//发起查询
	resp, err := http.Get(fmt.Sprintf("https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/%s?mchid=%s", orderNo, mchId))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("微信支付结果查询失败, statusCode:%d", resp.StatusCode)
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
