package ginlib

import (
	"context"
	"errors"
	"github.com/beego/i18n"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path"
	"strings"
)

func I18nInit(langPath string) {
	f, err := os.Open(langPath)
	if err != nil {
		panic(err)
	}
	fi, err := f.Readdir(10)
	if err != nil {
		panic(err)
	}
	for _, val := range fi {
		if val.IsDir() {
			continue
		}
		langName := strings.TrimSuffix(val.Name(), ".ini")
		langFile := path.Join(langPath, val.Name())
		if err = i18n.SetMessage(langName, langFile); err != nil {
			panic(err)
		} else {
			Logger.Debug("加载i18n文件成功", zap.String("lang", langName), zap.String("file", langFile))
		}
	}
}

// Lang 根据context获取lang
func Lang(ctx context.Context) (lang string) {
	if tmp := ctx.Value("lang"); tmp != nil {
		lang, _ = tmp.(string)
	}
	if lang == "" {
		lang = "en"
	}
	return
}

type ErrorI18n struct {
	i18nCode string
	args     []interface{}
}

// ErrorI18nNew 创建一个基于i18n的错误
func ErrorI18nNew(i18nCode string, args ...interface{}) ErrorI18n {
	return ErrorI18n{i18nCode, args}
}

func (e ErrorI18n) Error() string {
	return i18n.Tr("en", e.i18nCode, e.args...)
}

func (e ErrorI18n) I18nCode() string {
	return e.i18nCode
}

func (e ErrorI18n) ErrorWithLang(lang string) string {
	return i18n.Tr(lang, e.i18nCode, e.args...)
}

func (c *Context) JsonError(err error, code ...int) {
	var resp GinJsonResp
	resp.ErrorCode = 1
	if len(code) > 0 {
		resp.ErrorCode = code[0]
	}

	var i18nErr ErrorI18n
	if errors.As(err, &i18nErr) {
		resp.Lang = Lang(c.Request.Context())
		resp.MsgCode = i18nErr.i18nCode
		resp.ErrorMessage = i18nErr.ErrorWithLang(resp.Lang)
	} else {
		resp.ErrorMessage = err.Error()
	}

	c.JSON(http.StatusOK, resp)
}
