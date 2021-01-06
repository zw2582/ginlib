package weishen_gin_lib

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type Context struct {
	*gin.Context
}

func (this *Context) JsonSucc(data gin.H, msgs ...string)  {
	msg := ""
	if len(msgs) > 0 {
		msg = strings.Join(msgs, ";")
	}
	this.JsonReturn(0, data, msg)
}

func (this *Context) JsonFail(msg string, datas ... gin.H) {
	data := make(gin.H)
	if len(datas) > 0 {
		data = datas[0]
	}
	this.JsonReturn(1, data, msg)
}

func (this *Context) JsonReturn(code int, data gin.H, msg string) {
	this.JSON(http.StatusOK, gin.H{
		"code":code,
		"data":data,
		"msg":msg,
	})
}

func (this *Context) InputString(key string, defs ...string) string {
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
