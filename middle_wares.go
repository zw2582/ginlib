package ginlib

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

//AuthWare 认证拦截
func AuthWare(c *gin.Context)  {
	this := Context{Context:c}

	jwtToken := c.GetHeader("Authorization")
	if len(jwtToken) < 7{
		this.JsonReturn(5003, nil, "请先登录")
		this.Abort()
		return
	}
	jwtToken = jwtToken[7:]
	jwtSecret := Ini_Str("auth.jwt_secret")
	if uid, err := JwtAuthUid(jwtToken, jwtSecret); err != nil {
		Logger.Debug("登录失败", zap.Error(err), zap.String("jwtToken", jwtToken), zap.String("jwtSecret", jwtSecret))
		this.JsonReturn(5003, nil, "请先登录")
		this.Abort()
		return
	} else {
		this.Set("uid", int64(uid))
	}

	c.Next()
}


//Cors 处理跨域请求,支持options访问
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}
