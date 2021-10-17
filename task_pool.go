package ginlib

import (
	"go.uber.org/zap"
	"time"
)

type TaskPool struct {
	TaskIdle chan int
	WaitTime time.Duration //获取协程超时时间，不是任务执行超时时间，0.代表一直等待
}

//NewTaskPool 创建一个任务池
func NewTaskPool(maxTaskIdle int, waitTime ...time.Duration) TaskPool {
	wt := time.Duration(0)
	if len(waitTime) > 0 {
		wt = waitTime[0]
	}
	return TaskPool{make(chan int, maxTaskIdle), wt}
}

func (this *TaskPool) Do(msg interface{}, fn func(interface{})) {
	if this.WaitTime > 0 {
		//定义获取协程超时时间
		tw := time.NewTimer(this.WaitTime)
		//能否获取到开启协程权限
		select {
		case this.TaskIdle <- 1:
		case <-tw.C:
			if Logger != nil {
				Logger.Info("获取可执行协程超时", zap.Duration("wait_time", this.WaitTime))
			}
			return
		}
	} else {
		//一直等待
		this.TaskIdle <- 1
	}

	//开启协程执行任务
	go func() {
		defer func() {
			//处理完毕后还回协程
			select {
			case <- this.TaskIdle:
			default:
				if Logger != nil {
					Logger.Error("处理完毕后还回协程池失败")
				}
			}
			//兜底错误处理
			if err := recover(); err != nil {
				switch errVal := err.(type) {
				case error:
					if Logger != nil {
						Logger.Error("开启协程执行任务捕获panic异常", zap.Error(errVal), zap.Stack("trace"))
					}
				default:
					if Logger != nil {
						Logger.Error("开启协程执行任务捕获panic异常", zap.Any("err", errVal), zap.Stack("trace"))
					}
				}

			}
		}()
		fn(msg)
	}()
}
