package ginlib

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type Context struct {
	*gin.Context
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