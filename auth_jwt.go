package ginlib

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
	"strconv"
	"time"
)

//认证user
type JwtAuth struct {
	Token string	//认证token
	JwtSecret string //jwt加密秘钥
	ExpireDration time.Duration //token有效时间, 默认30天后失效
	uid int
	authed int //是否token认证过 1.已认证 0.未认证
	autherr error
}

func (this *JwtAuth) GetJwtSecret() string {
	if this.JwtSecret != "" {
		return this.JwtSecret
	}
	//默认secret
	return "1611c7e06040bb0a7c549ee2a17d42ef"
}

//Uid 用户id
func (this *JwtAuth) Uid() (int, error) {
	if this.authed == 1 {
		return this.uid, this.autherr
	}
	if this.Token == "" {
		this.autherr = fmt.Errorf("登录已失效")
		return this.uid, this.autherr
	}

	//校验token
	claim := jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(this.Token, &claim, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(this.GetJwtSecret()), nil
	})
	if err != nil {
		Logger.Error(err.Error(), zap.Error(err))
		this.autherr = fmt.Errorf("登录已失效")
		return this.uid, this.autherr
	}
	if !token.Valid {
		this.autherr = fmt.Errorf("登录已失效")
		return this.uid, this.autherr
	}
	id,_ := strconv.Atoi(claim.Id)
	if id == 0 {
		this.autherr = fmt.Errorf("登录已失效")
		return this.uid, this.autherr
	}
	this.uid = id
	this.authed = 1
	return  this.uid, nil
}

//JwtRefresh 刷新
func (this *JwtAuth) JwtRefresh() (string, error) {
	uid, err := this.Uid()
	if err != nil {
		return "", err
	}

	//产生新的token
	return this.JwtLogin(uid)
}

//JwtLogin 登录
func (this *JwtAuth) JwtLogin(uid int) (tokenStr string, err error) {
	if uid == 0 {
		return "", fmt.Errorf("用户不存在")
	}
	dration := this.ExpireDration
	if dration == 0 {
		dration = time.Hour*24*30 //30天后失效
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(dration).Unix(),
		Subject:"",
		Id:strconv.Itoa(uid),
	})

	tokenStr, err = token.SignedString([]byte(this.GetJwtSecret()))
	if err != nil {
		panic(err)
	}
	this.uid = uid
	this.autherr = nil
	this.authed = 1
	this.Token = tokenStr
	return
}

