package ginlib

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	Logger *zap.Logger
)

//InitLogger 初始化日志文件
func InitLogger(rotateSig ...syscall.Signal) *zap.Logger {
	log.Println("初始化日志文件")
	//log.path 使用","表示多个日志文件，如果日志文件是system，则走系统输出
	logpath := Ini_Str("log.path")
	loglevel := Ini_Str("log.level")
	compress := Ini_Bool("log.compress", false)

	Logger = CreateLogger(logpath, loglevel, compress, rotateSig...)

	return Logger
}

//CreateLogger 创建zaplogger
func CreateLogger(logpath, loglevel string, compress bool, rotateSig ...syscall.Signal) *zap.Logger {
	logPaths := strings.Split(logpath, ",")
	hooks := make([]zapcore.WriteSyncer, 0)
	for _, val := range logPaths {
		val = strings.Trim(val, " ")
		if val == "" {
			continue
		}
		if val == "system" {
			hooks = append(hooks, zapcore.AddSync(os.Stdout))
		} else {
			hook := lumberjack.Logger{
				Filename:   val,  //日志文件路径
				MaxSize:    128, //最大字节
				MaxAge:     30,
				MaxBackups: 7,
				Compress:   compress,
				LocalTime: true,
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


