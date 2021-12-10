package ginlib

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

func GoRecover(fun func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch val := e.(type) {
			case error:
				err = val
				Logger.Error("GoRecover捕获panic", zap.Error(val), zap.Stack("GoRecover"))
			default:
				err = fmt.Errorf("%+v\n", val)
				Logger.Error("GoRecover捕获panic", zap.Any("err", val), zap.Stack("GoRecover"))
			}
		}
	}()
	fun()
	return nil
}

// GinRecovery 接受gin框架http中panic的错误
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					Logger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					// 增加dingtalk
					NoticeDingtalk(gin.H{
						"error":   fmt.Sprintf("%v", err),
						"request": string(httpRequest),
						"trace":   string(debug.Stack()),
					})
					Logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.StackSkip("track", 1),
						//zap.String("stack", string(debug.Stack())),
					)
				} else {
					// 增加dingtalk
					NoticeDingtalk(gin.H{
						"error":   fmt.Sprintf("%v", err),
						"request": string(httpRequest),
						"trace":   string(debug.Stack()),
					})
					Logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

// GinLogger 接收gin框架的http请求日志，通常不使用
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		Logger.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

//NoticeDingtalk 钉钉通知严重错误
func NoticeDingtalk(obj gin.H) (bs []byte, err error) {
	contentByte, _ := json.MarshalIndent(obj, "", "\t")
	content := strings.Replace(string(contentByte), "\"", "", -1)
	s := `{"msgtype": "text","text": {"content": "panic:` + APP_NAME + "[" + GetEnv() + "]\n" + content + `"}}`

	dingAccessToken := Ini_Str("ding.access_token")
	Secret := Ini_Str("ding.secret")
	if dingAccessToken == "" {
		Logger.Debug("请先配置钉钉 ding.access_token")
		return
	}
	if Secret == "" {
		Logger.Debug("请先配置钉钉 ding.secret")
		return
	}

	client := &http.Client{}
	urlStr := "https://oapi.dingtalk.com/robot/send?access_token="+dingAccessToken
	method := "POST"

	//  构建 签名
	//  把timestamp+"\n"+密钥当做签名字符串，使用HmacSHA256算法计算签名，然后进行Base64 encode，最后再把签名参数再进行urlEncode，得到最终的签名（需要使用UTF-8字符集）。
	timeStampNow := time.Now().UnixNano() / 1000000
	signStr := fmt.Sprintf("%d\n%s", timeStampNow, Secret)

	hash := hmac.New(sha256.New, []byte(Secret))
	hash.Write([]byte(signStr))
	sum := hash.Sum(nil)

	encode := base64.StdEncoding.EncodeToString(sum)
	urlEncode := url.QueryEscape(encode)

	// 构建 请求 urlStr
	urlStr = fmt.Sprintf("%s&timestamp=%d&sign=%s", urlStr, timeStampNow, urlEncode)

	payload := strings.NewReader(s)
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	bs, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	return
}

//Context 自定义包装gin上下文
type Context struct {
	*gin.Context
}

//Uid 获取用户uid
func (this *Context) Uid() int64 {
	return this.GetInt64("uid")
}

func (this *Context) JsonSucc(data interface{}, msgs ...string)  {
	msg := ""
	if len(msgs) > 0 {
		msg = strings.Join(msgs, ";")
	}
	this.JsonReturn(0, data, msg)
}

func (this *Context) JsonFail(msg string, datas ... interface{}) {
	var data interface{}
	if len(datas) > 0 {
		data = datas[0]
	}
	this.JsonReturn(1, data, msg)
}

func (this *Context) JsonReturn(code int, data interface{}, msg string) {
	this.JSON(http.StatusOK, gin.H{
		"code":code,
		"data":data,
		"msg":msg,
	})
}

func (this *Context) InputStr(key string, defs ...string) string {
	def := ""
	if len(defs) > 0 {
		def = defs[0]
	}
	val := this.Query(key)
	if val == "" {
		val = this.PostForm(key)
	}
	if val == "" {
		return def
	}
	return val
}

func (this *Context) InputInt(key string, defs ...int) int {
	def := 0
	if len(defs) > 0 {
		def = defs[0]
	}
	val := this.Query(key)
	if val == "" {
		val = this.PostForm(key)
	}
	if val == "" {
		return def
	}
	t,_ := strconv.Atoi(val)

	return t
}

func (this *Context) InputInt32(key string, defs ...int32) int32 {
	def := int32(0)
	if len(defs) > 0 {
		def = defs[0]
	}
	val := this.Query(key)
	if val == "" {
		val = this.PostForm(key)
	}
	if val == "" {
		return def
	}
	t,_ := strconv.ParseInt(val, 10, 32)

	return int32(t)
}

func (this *Context) InputInt64(key string, defs ...int64) int64 {
	def := int64(0)
	if len(defs) > 0 {
		def = defs[0]
	}
	val := this.Query(key)
	if val == "" {
		val = this.PostForm(key)
	}
	if val == "" {
		return def
	}
	t,_ := strconv.ParseInt(val, 10, 64)

	return t
}

func (this *Context) InputFloat32(key string, defs ...float32) float32 {
	var def float32
	if len(defs) > 0 {
		def = defs[0]
	}
	val := this.Query(key)
	if val == "" {
		val = this.PostForm(key)
	}
	if val == "" {
		return def
	}
	t,_ := strconv.ParseFloat(val, 32)

	return float32(t)
}

func (this *Context) InputFloat64(key string, defs ...float64) float64 {
	var def float64
	if len(defs) > 0 {
		def = defs[0]
	}
	val := this.Query(key)
	if val == "" {
		val = this.PostForm(key)
	}
	if val == "" {
		return def
	}
	t,_ := strconv.ParseFloat(val, 64)

	return t
}

func (this *Context) InputBool(key string, defs ...bool) bool {
	var def bool
	if len(defs) > 0 {
		def = defs[0]
	}
	val := this.Query(key)
	if val == "" {
		val = this.PostForm(key)
	}
	if val == "" {
		return def
	}
	t,_ := strconv.ParseBool(val)

	return t
}

//InnerSucc 内部服务返回成功
func (this *Context) InnerSucc(data interface{}) {
	this.InnerReturn(100, data, "")
}

//InnerFail 内部服务返回失败信息
func (this *Context) InnerFail(msg string) {
	this.InnerReturn(101, nil, msg)
}

//InnerReturn 内部服务返回
// code 100:代表成功，其他的自定义错误，必须是3位数的code
func (this *Context) InnerReturn(code int, data interface{}, msg string) {
	if code == 100 {
		//当100成功时，会将数据转换为json传递给msg
		tmp, _ := json.Marshal(data)
		msg = string(tmp)
	}
	this.String(http.StatusOK, "%d%s", msg)
}
