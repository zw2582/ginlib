package ginlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

var (
	Logger *zap.Logger
)

// InitLogger 初始化日志文件
func InitLogger(rotateSig ...syscall.Signal) *zap.Logger {
	log.Println("初始化日志文件")
	//log.path 使用","表示多个日志文件；stdout:输出到stdout
	logPath := Ini_Str("log.path", "stdout")
	//level等级 debug、info、error 默认debug
	loglevel := Ini_Str("log.level", "debug")
	//公用编码器 json:json格式输出，不配置:普通格式输出
	logEncode := Ini_Str("log.encode")
	// 是否压缩日志
	logCompress := Ini_Bool("log.compress", false)

	Logger = CreateLogger(logPath, loglevel, logEncode, logCompress, rotateSig...)

	return Logger
}

// CreateLogger 创建zapLogger
func CreateLogger(logPath, loglevel, logEncode string, logCompress bool, rotateSig ...syscall.Signal) *zap.Logger {
	logPaths := strings.Split(logPath, ",")
	hooks := make([]zapcore.WriteSyncer, 0)
	for _, val := range logPaths {
		val = strings.Trim(val, " ")
		if val == "" {
			continue
		}
		if val == "stdout" {
			hooks = append(hooks, zapcore.AddSync(os.Stdout))
		} else {
			hook := lumberjack.Logger{
				Filename:   path.Join(val, Ini_Str("app.name")+".log"), //日志文件路径
				MaxSize:    128,                                        //最大字节
				MaxAge:     30,
				MaxBackups: 7,
				Compress:   logCompress,
				LocalTime:  true,
			}
			// 接收信号切割
			if len(rotateSig) > 0 {
				sigs := make(chan os.Signal, 1)
				signal.Notify(sigs, rotateSig[0])
				go func() {
					for _ = range sigs {
						hook.Rotate()
					}
				}()
			}
			hooks = append(hooks, zapcore.AddSync(&hook))
		}
	}

	//没有日志写入钩子时，使用默认系统输出
	if len(hooks) == 0 {
		hooks = append(hooks, zapcore.AddSync(os.Stdout))
	}
	w := zapcore.NewMultiWriteSyncer(hooks...)
	// 设置日志级别,debug可以打印出info,debug,warn；info级别可以打印warn，info；warn只能打印warn
	// debug->info->warn->error
	var level zapcore.Level
	switch loglevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	//公用编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,                            // 大写编码器
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"), // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,                         //
		EncodeCaller:   zapcore.ShortCallerEncoder,                             // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	zapEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	if logEncode == "json" {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		zapEncoder = zapcore.NewJSONEncoder(encoderConfig)
	}
	core := zapcore.NewCore(zapEncoder, w, level)

	return zap.New(core, zap.AddCaller())
}

var (
	larkLastTime = make(map[string]time.Time)
)

// LarkNotice 通知消息
func LarkNotice(target string, texts ...string) {
	targetHooks := map[string]string{
		"system": "https://open.larksuite.com/open-apis/bot/v2/hook/94341b5b-0f3c-4f18-9e3b-d794973563cc",
		"data":   "https://open.larksuite.com/open-apis/bot/v2/hook/94341b5b-0f3c-4f18-9e3b-d794973563cc",
	}
	targetTitles := map[string]string{
		"system": "程序异常报警",
		"data":   "数据异常报警",
	}
	botHook := targetHooks[target]
	if botHook == "" {
		Logger.Error("无该target的hook地址", zap.String("target", target))
		return
	}
	contents := make([]interface{}, 0)
	texts = append([]string{fmt.Sprintf("项目环境: %s", GetEnv())}, texts...)
	for _, val := range texts {
		contents = append(contents, []gin.H{
			{
				"tag":  "text",
				"text": val,
			},
		})
	}
	//一分钟内只触发一次
	if time.Now().Sub(larkLastTime[target]) < time.Minute {
		return
	}
	larkLastTime[target] = time.Now()
	param := gin.H{
		"msg_type": "post",
		"content": gin.H{
			"post": gin.H{
				"zh_cn": gin.H{
					"title":   targetTitles[target],
					"content": contents,
				},
			},
		},
	}
	paramJson, _ := json.Marshal(param)
	req, err := http.NewRequest("POST", botHook, bytes.NewReader(paramJson))
	if err != nil {
		Logger.Error("通知lark失败", zap.Error(err))
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		Logger.Error("通知lark失败", zap.Error(err))
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	Logger.Debug("发送lark通知", zap.ByteString("raw", raw), zap.String("botHook", botHook))
}
