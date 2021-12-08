package ginlib

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

//TickerTaskService 该服务用于异步定时执行一些方法，同时，如果任务的上一个周期没有执行结束，本次定时即使到了也不会执行
type TickerTaskService struct {
	taskList []*TickerTask
}

//AddTask 创建任务
func (this *TickerTaskService) AddTask(name string, d time.Duration, f TickerTaskFun) {
	t := &TickerTask{
		d:d,
		f:f,
		name:name,
		doing:false,
	}
	this.taskList = append(this.taskList, t)
}

//TaskSingleStart 开始单任务执行启动
func (this *TickerTaskService)TaskSingleStart() {
	Logger.Info("启动单任务处理定时任务服务...")
	for _,task := range this.taskList {
		task.execSingle()
	}
	//定义秒级定时器
	t := time.NewTicker(time.Second)

	for _ = range t.C {
		for _,task := range this.taskList {
			if time.Now().Sub(task.lastTime) >= task.d {
				task.execSingle()
			}
		}
	}
}

//TaskSingleStart2 将TaskSingleStart的ticker实现改为sleep实现，对比效果
func (this *TickerTaskService)TaskSingleStart2() {
	for {
		for _,task := range this.taskList {
			if time.Now().Sub(task.lastTime) >= task.d {
				task.execSingle()
			}
		}
		time.Sleep(time.Second)
	}
}

//TickerTaskFun 执行任务的方法
type TickerTaskFun func()

//TickerTask 任务
type TickerTask struct {
	lastTime time.Time
	d        time.Duration
	f        TickerTaskFun
	name     string
	doing    bool //是否正在执行
	lock     sync.Mutex
}

//execSingle 单任务异步执行任务
func (this *TickerTask) execSingle() {
	Logger.Debug(fmt.Sprintf("task:%s 定时器监测到需要执行,lastTime:%s", this.name, this.lastTime.Format("2006-01-02 15:04:05")))

	//排除重复执行
	this.lock.Lock()
	if this.doing {
		Logger.Debug(fmt.Sprintf("task:%s 存在进程正在执行，本次执行中断", this.name))
		this.lock.Unlock()
		return
	}
	this.doing = true
	this.lock.Unlock()

	this.lastTime = time.Now()

	//异步执行任务
	go func() {
		//定义异常记录
		defer func() {
			if err := recover(); err != nil {
				Logger.Error(fmt.Sprintf("task:%s 异常:%s trace:%s", this.name, err, debug.Stack()))
			}
			this.lock.Lock()
			this.doing = false
			this.lock.Unlock()
		}()
		//执行任务
		Logger.Debug(fmt.Sprintf("task:%s 开始执行任务", this.name))
		this.f()
	}()

}