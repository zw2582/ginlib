package ginlib

import (
	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"
)

var (
	Logger *zap.Logger
)

//InitLogger 初始化日志文件
func InitLogger(rotateSig ...syscall.Signal) *zap.Logger {
	log.Println("初始化日志文件")
	logpath := Ini_Str("log.path", "./logs/project.log")
	loglevel := Ini_Str("log.level")
	compress := Ini_Bool("log.compress", false)

	Logger = CreateLogger(logpath, loglevel, compress, rotateSig[0])

	return Logger
}

//CreateLogger 创建zaplogger
func CreateLogger(logpath, loglevel string, compress bool, rotateSig syscall.Signal) *zap.Logger {
	hook := lumberjack.Logger{
		Filename:   logpath,  //日志文件路径
		MaxSize:    128, //最大字节
		MaxAge:     30,
		MaxBackups: 7,
		Compress:   compress,
		LocalTime: true,
	}

	// 接收信号切割
	if rotateSig > 0 {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, rotateSig)
		go func() {
			for _ = range sigs {
				hook.Rotate()
			}
		}()
	}
	w := zapcore.AddSync(&hook)
	env := Ini_Str("app.env")
	if env == "dev" {
		//开发环境打印控制台
		w = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook))
	}

	// 设置日志级别,debug可以打印出info,debug,warn；info级别可以打印warn，info；warn只能打印warn
	// debug->info->warn->error
	var level zapcore.Level
	switch loglevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level= zap.InfoLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	//公用编码器
	encodeStyle := Ini_Str("log.encode", "console")
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,  // 大写编码器
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.ShortCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	zapEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	if encodeStyle == "json" {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		zapEncoder = zapcore.NewJSONEncoder(encoderConfig)
	}
	core := zapcore.NewCore(zapEncoder, w, level)

	log := zap.New(core, zap.AddCaller())
	return log
}

// GinLogger 接收gin框架默认的日志
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

// GinRecovery recover掉项目可能出现的panic，并使用zap记录相关日志
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
					Logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
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