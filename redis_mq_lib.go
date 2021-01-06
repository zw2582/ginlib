package ginlib

import (
	"github.com/go-redis/redis"
)

//redis消息队列
type MqRedisSer struct {
	QueueName string //消息队列名称
	QueueBackName string //防丢队列名称
	RedisCli *redis.Client	//redis连接客户端
}

//PushMessage 塞入消息
func (this *MqRedisSer) PushMessage(msg string) int64 {
	if msg == "" {
		return 0
	}

	return this.RedisCli.LPush(this.QueueName, msg).Val()
}

//GetMessage 获取消息
func (this *MqRedisSer) GetMessage() string {
	res := this.RedisCli.RPopLPush(this.QueueName, this.QueueBackName)
	return res.Val()
}

//MessageAck 消息
func (this *MqRedisSer) MessageAck(msg string) int64 {
	if msg == "" {
		return 0
	}
	val := this.RedisCli.LRem(this.QueueBackName, -1, msg).Val()

	//清除待修复数据
	repeatKey := this.QueueName+"_repeat"
	md5res := Md5encode(msg)
	this.RedisCli.HDel(repeatKey, md5res)

	return val
}

//Repeat 回滚异常队列
func (this *MqRedisSer) Repeat() int {
	n := 0
	repeatKey := this.QueueName+"_repeat"
	for {
		res := this.RedisCli.RPop(this.QueueBackName).Val()
		if res == "" {
			break
		}

		md5res := Md5encode(res)
		cnt := this.RedisCli.HIncrBy(repeatKey, md5res, 1).Val()
		if cnt > 3 {
			this.RedisCli.HDel(repeatKey, md5res)
			continue
		}
		this.RedisCli.LPush(this.QueueName, res)
		n++
	}
	return n
}
