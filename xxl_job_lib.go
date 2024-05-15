package ginlib

import (
	"fmt"
	xxl "github.com/xxl-job/xxl-job-executor-go"
	"go.uber.org/zap"
)

// XXLJobCreate 创建xxl job
func XXLJobCreate(logHandle xxl.LogHandler) (exec xxl.Executor, err error) {
	xxlAddr := Ini_Str("xxl.addr")   //xxl server的地址
	xxlToken := Ini_Str("xxl.token") //请求令牌(默认为空)
	xxlKey := Ini_Str("xxl.key")     //执行器名称
	if xxlAddr == "" || xxlKey == "" {
		return exec, fmt.Errorf("请配置xxl.addr等信息")
	}
	//初始化执行器
	exec = xxl.NewExecutor(
		xxl.ServerAddr(xxlAddr),
		xxl.AccessToken(xxlToken),   //请求令牌(默认为空)
		xxl.ExecutorPort(APP_PORT),  //默认9999（非必填）
		xxl.RegistryKey(xxlKey),     //执行器名称
		xxl.SetLogger(&XXLLogger{}), //自定义日志
	)
	exec.Init()
	//设置日志查看handler
	exec.LogHandler(logHandle)

	Logger.Info("初始化xxlJob", zap.String("Addr", xxlAddr), zap.String("Token", xxlToken), zap.String("Key", xxlKey))
	return exec, nil
}

// XXLLogger xxl.Logger接口实现
type XXLLogger struct{}

func (l *XXLLogger) Info(format string, a ...interface{}) {
	Logger.Info(fmt.Sprintf("[xxl.Logger] - "+format, a...))
}

func (l *XXLLogger) Error(format string, a ...interface{}) {
	Logger.Debug(fmt.Sprintf("[xxl.Logger] - "+format, a...))
}
