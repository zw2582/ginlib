package ginlib

import (
	"go.uber.org/zap"
	"time"
)

//GoCoveryLoop 不死的捕获子协程开启
func GoCoveryLoop(title string, fun func()) {
	go func() {
		for {
			func() {
				defer func() {
					//捕获panic，并记录
					if e := recover(); e != nil {
						switch val := e.(type) {
						case error:
							Logger.Fatal("GoCoveryLoop捕获panic", zap.String("title",title), zap.Error(val), zap.Stack("GoCoveryLoop"))
						default:
							Logger.Fatal("GoCoveryLoop捕获panic", zap.String("title",title), zap.Any("err", val), zap.Stack("GoCoveryLoop"))
						}
					}
				}()
				//执行调用函数
				Logger.Info("异步任务执行开始:"+title)
				fun()
			}()
			//一旦发生失败，停顿5秒继续执行
			time.Sleep(time.Second * 5)
		}
	}()
}
