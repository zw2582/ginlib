package ginlib

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
	"time"
)

//redis消息队列
type RedisMq struct {
	queueName     string        //消息队列名称
	queueBackName string        //防丢队列名称
	hashRepeatCntName string //记录修复次数，修复超过3次，则不在进入防丢队列
	redisCli      *redis.Client //redis连接客户端
}

//NewRedisMq 新建一个redis消息队列
func NewRedisMq(queueName string, redisCli *redis.Client) RedisMq {
	if redisCli == nil {
		panic(errors.New("请为RedisMq传递有效的redis客户端"))
	}
	if err := redisCli.Ping().Err(); err != nil {
		panic(err)
	}
	cli := RedisMq{
		queueName:     queueName,
		queueBackName: fmt.Sprintf("%s_backend_mq_35862714", queueName),
		hashRepeatCntName: fmt.Sprintf("%s_repeat_mq_35862714", queueName),
		redisCli:      redisCli,
	}

	//每隔一段时间修复消息队列,filterRepeatKey 用于多个进程同时修复时，只有一个进程有修复权限
	filterRepeatKey := fmt.Sprintf("%s_filter", cli.hashRepeatCntName)
	GoCoveryLoop("间隔5分钟修复消息队列:"+cli.hashRepeatCntName, func() {
		d := time.Minute * 5 //修复间隔时间
		//启动先修复一次
		if redisCli.SetNX(filterRepeatKey, 1, d - time.Second).Val() {
			Logger.Info("我拿到修复消费队列权限:"+cli.hashRepeatCntName)
			cli.Repeat()
		}
		//定时修复
		tk := time.NewTicker(d)
		for _ = range tk.C {
			if redisCli.SetNX(filterRepeatKey, 1, d - time.Second).Val() {
				Logger.Info("我拿到修复消费队列权限:"+cli.hashRepeatCntName)
				cli.Repeat()
			}
		}
	})

	return cli
}

//PushMessage 塞入消息
func (this *RedisMq) PushMessage(msg string) int64 {
	if msg == "" {
		return 0
	}

	if val, err := this.redisCli.LPush(this.queueName, msg).Result(); err != nil {
		panic(err)
	} else {
		return val
	}
}

//GetMessage 获取消息
//	blockDration 阻塞一段时间获取
func (this *RedisMq) GetMessage(blockDration ...time.Duration) string {
	if len(blockDration) == 0 {
		res := this.redisCli.RPopLPush(this.queueName, this.queueBackName)
		return res.Val()
	} else {
		res := this.redisCli.BRPopLPush(this.queueName, this.queueBackName, blockDration[0])
		return res.Val()
	}
}

//MessageAck 消息
func (this *RedisMq) MessageAck(msg string) int64 {
	if msg == "" {
		return 0
	}
	val := this.redisCli.LRem(this.queueBackName, -1, msg).Val()

	//清除待修复数据
	repeatKey := this.hashRepeatCntName
	md5res := Md5encode(msg)
	this.redisCli.HDel(repeatKey, md5res)

	return val
}

//Repeat 回滚异常队列
func (this *RedisMq) Repeat() int64 {
	n := int64(0)
	repeatKey := this.hashRepeatCntName
	//备份队列尾部插入修复标志符号
	this.redisCli.LSet(this.queueBackName, 0, "x0x")
	//停止60秒，等待客户端ack
	time.Sleep(time.Minute)
	//循环pop备份数据，并修复
	backLen := this.redisCli.LLen(this.queueBackName).Val()
	for ; n< backLen; n++ {
		res := this.redisCli.RPop(this.queueBackName).Val()
		if res == "" || res == "x0x" {
			break
		}

		md5res := Md5encode(res)
		cnt := this.redisCli.HIncrBy(repeatKey, md5res, 1).Val()
		if cnt > 3 {
			Logger.Info("修复超过3次，放弃修复", zap.String("res", res))
			this.redisCli.HDel(repeatKey, md5res)
			continue
		}
		Logger.Info("修复数据", zap.String("res", res))
		this.redisCli.LPush(this.queueName, res)
	}
	return n
}
