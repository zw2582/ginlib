package abo

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

var (
	Logger *zap.Logger
)

func loggerDefaultInit(projectName string) {
	logFile := ""
	logLevel := "debug"
	logEncode := "console"
	if apolloConfigDefault != nil {
		logPath := ConfigVal("log.path")
		if logPath != "" {
			logPaths := strings.Split(logPath, ";")
			logFiles := make([]string, 0)
			for _, val := range logPaths {
				if val == "stdout" {
					logFiles = append(logFiles, val)
				} else {
					logFiles = append(logFiles, path.Join(val, projectName+".log"))
				}
			}
			if len(logFiles) > 0 {
				logFile = strings.Join(logFiles, ",")
			}
		}
		logLevel = ConfigStr("log.level", logLevel)
		logEncode = ConfigVal("log.encode")
	}
	//默认打印控制台
	if logFile == "" {
		logFile = "stdout"
	}
	fmt.Println("初始化日志", logFile, logLevel, logEncode)

	Logger = CreateLogger(logFile, logLevel, logEncode, true)
}

// CreateLogger 创建zap logger
// logPath 使用","表示多个日志文件，如果日志文件是system，则走系统输出
func CreateLogger(logPath, logLevel, logEncode string, compress bool, rotateSig ...syscall.Signal) *zap.Logger {
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
				Filename:   val, //日志文件路径
				MaxSize:    128, //最大字节
				MaxAge:     30,
				MaxBackups: 7,
				Compress:   compress,
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
	switch logLevel {
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

	log := zap.New(core, zap.AddCaller())
	return log
}
