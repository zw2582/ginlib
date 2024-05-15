package ginlib

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Context 自定义包装gin上下文
type Context struct {
	*gin.Context
}

type GinJsonResp struct {
	ErrorCode    int         `json:"error_code"`
	ErrorMessage string      `json:"error_message"`
	Data         interface{} `json:"data"`
	MsgCode      string      `json:"msg_code"`
	Lang         string      `json:"lang"`
}

// User 获取用户
func (c *Context) User() (user interface{}) {
	if val, ok := c.Get("user"); ok {
		return val
	}
	return
}

func (c *Context) UserSet(user interface{}) {
	c.Set("user", user)
}

func (c *Context) JsonSucc(data interface{}, msgs ...string) {
	msg := ""
	if len(msgs) > 0 {
		msg = strings.Join(msgs, ";")
	}
	c.JsonReturn(0, data, msg)
}

func (c *Context) JsonFail(msg string, datas ...interface{}) {
	var data interface{}
	if len(datas) > 0 {
		data = datas[0]
	}
	c.JsonReturn(1, data, msg)
}

func (c *Context) JsonReturn(code int, data interface{}, msg string) {
	c.JSON(http.StatusOK, GinJsonResp{
		ErrorCode:    code,
		ErrorMessage: msg,
		Data:         data,
		MsgCode:      "",
		Lang:         "",
	})
}

func (c *Context) InputStr(key string, defs ...string) string {
	def := ""
	if len(defs) > 0 {
		def = defs[0]
	}
	val := c.Query(key)
	if val == "" {
		val = c.PostForm(key)
	}
	if val == "" {
		return def
	}
	return val
}

func (c *Context) InputInt(key string, defs ...int) int {
	def := 0
	if len(defs) > 0 {
		def = defs[0]
	}
	val := c.Query(key)
	if val == "" {
		val = c.PostForm(key)
	}
	if val == "" {
		return def
	}
	t, _ := strconv.Atoi(val)

	return t
}

func (c *Context) InputInt32(key string, defs ...int32) int32 {
	def := int32(0)
	if len(defs) > 0 {
		def = defs[0]
	}
	val := c.Query(key)
	if val == "" {
		val = c.PostForm(key)
	}
	if val == "" {
		return def
	}
	t, _ := strconv.ParseInt(val, 10, 32)

	return int32(t)
}

func (c *Context) InputInt64(key string, defs ...int64) int64 {
	def := int64(0)
	if len(defs) > 0 {
		def = defs[0]
	}
	val := c.Query(key)
	if val == "" {
		val = c.PostForm(key)
	}
	if val == "" {
		return def
	}
	t, _ := strconv.ParseInt(val, 10, 64)

	return t
}

func (c *Context) InputFloat32(key string, defs ...float32) float32 {
	var def float32
	if len(defs) > 0 {
		def = defs[0]
	}
	val := c.Query(key)
	if val == "" {
		val = c.PostForm(key)
	}
	if val == "" {
		return def
	}
	t, _ := strconv.ParseFloat(val, 32)

	return float32(t)
}

func (c *Context) InputFloat64(key string, defs ...float64) float64 {
	var def float64
	if len(defs) > 0 {
		def = defs[0]
	}
	val := c.Query(key)
	if val == "" {
		val = c.PostForm(key)
	}
	if val == "" {
		return def
	}
	t, _ := strconv.ParseFloat(val, 64)

	return t
}

func (c *Context) InputBool(key string, defs ...bool) bool {
	var def bool
	if len(defs) > 0 {
		def = defs[0]
	}
	val := c.Query(key)
	if val == "" {
		val = c.PostForm(key)
	}
	if val == "" {
		return def
	}
	t, _ := strconv.ParseBool(val)

	return t
}

// InnerSucc 内部服务返回成功
func (c *Context) InnerSucc(data interface{}) {
	c.InnerReturn(100, data, "")
}

// InnerFail 内部服务返回失败信息
func (c *Context) InnerFail(msg string) {
	c.InnerReturn(101, nil, msg)
}

// InnerReturn 内部服务返回
// code 100:代表成功，其他的自定义错误，必须是3位数的code
func (c *Context) InnerReturn(code int, data interface{}, msg string) {
	if code == 100 {
		//当100成功时，会将数据转换为json传递给msg
		tmp, _ := json.Marshal(data)
		msg = string(tmp)
	}
	c.String(http.StatusOK, "%d%s", msg)
}

// GracefulExitWeb 具备优雅停止web服务的启动方式
func GracefulExitWeb(engine *gin.Engine, host, port string) {
	log.Println("web服务启动, listen:" + host + ":" + port)
	srv := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: engine.Handler(),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("server listen failed! error:", err.Error())
		}
	}()

	// 平滑重启
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	sig := <-ch

	log.Println("got a signal", sig)
	now := time.Now()
	cxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := srv.Shutdown(cxt)
	if err != nil {
		log.Println("err", err)
	}

	// 看看实际退出所耗费的时间
	log.Println("------exited--------", time.Since(now))
}

// GinRecovery 接受gin框架http中panic的错误
func GinRecovery(stack, notice bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if e := recover(); e != nil {
				var err error
				switch v := e.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("%v", v)
				}

				//记录日志
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if stack {
					Logger.Error("[Recovery from panic] path:"+c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.StackSkip("track", 1),
					)
				} else {
					Logger.Error("[Recovery from panic] path:"+c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}

				//返回系统错误
				this := Context{c}
				this.JsonError(ErrorI18nNew("系统异常"))
				c.Abort()

				// lark通知
				if notice {
					go LarkNotice("system", fmt.Sprintf("错误描述:%s", err.Error()), fmt.Sprintf("堆栈:\n%s", string(debug.Stack())), fmt.Sprintf("请求体:\n%s", string(httpRequest)))
				}
				return
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

var (
	//MetricRequestDuration 请求耗时
	MetricRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration",
		Buckets: []float64{10, 20, 30, 40, 50, 60, 70, 100, 200, 500},
	}, []string{"path", "code"})
)

// MiddleRequestMetric 请求指标计算
func MiddleRequestMetric(c *gin.Context) {
	start := time.Now()
	c.Next()
	MetricRequestDuration.WithLabelValues(c.Request.URL.Path, strconv.Itoa(c.Writer.Status())).Observe(float64(time.Since(start).Milliseconds()))
}
