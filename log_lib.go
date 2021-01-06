package ginlib

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
)

var (
	Logger *zap.Logger
)

//InitLogger 初始化日志文件
func InitLogger() *zap.Logger {
	log.Println("初始化日志文件")

	logpath := Ini_Str("log.path", "./logs/project.log")

	loglevel := Ini_Str("log.level")

	Logger = CreateLogger(logpath, loglevel)

	return Logger
}

//CreateLogger 创建zaplogger
func CreateLogger(logpath, loglevel string) *zap.Logger {
	hook := lumberjack.Logger{
		Filename:   logpath,  //日志文件路径
		MaxSize:    128, //最大字节
		MaxAge:     30,
		MaxBackups: 7,
		Compress:   true,
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
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,  // 大写编码器
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