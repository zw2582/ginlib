package ginlib

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"strconv"
	"time"
)

//JwtAuthUid 解析用户id
func JwtAuthUid(jwtToken, jwtSecret string) (uid int, err error) {
	//校验token
	claim := jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(jwtToken, &claim, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
			return nil, fmt.Errorf("无法解析jwttoken,其算法: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return
	}
	if !token.Valid {
		err = fmt.Errorf("登录已失效")
		return
	}
	uid,_ = strconv.Atoi(claim.Id)
	if uid == 0 {
		err = fmt.Errorf("登录已失效")
		return
	}
	return
}

//JwtAuthLogin 使用jwt登录获得认证token
func JwtAuthLogin(uid int, jwtSecret string, duration time.Duration) (tokenStr string, err error) {
	if uid == 0 {
		return "", fmt.Errorf("用户不存在")
	}
	if duration == 0 {
		duration = time.Hour*24*30 //30天后失效
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(duration).Unix(),
		Subject:"",
		Id:strconv.Itoa(uid),
	})

	tokenStr, err = token.SignedString([]byte(jwtSecret))
	if err != nil {
		return
	}
	return
}

