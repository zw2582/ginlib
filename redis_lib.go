package ginlib

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
	"log"
	"net"
	"reflect"
	"time"
)

var (
	RedisCli *redis.Client
)

//InitRedis 初始化redis
func InitRedis()  {
	log.Println("初始化redis")

	host := Ini_Str("redis.host", "127.0.0.1")
	port := Ini_Str("redis.port",  "6379")
	db := Ini_Int("redis.database")
	password := Ini_Str("redis.password")
	//建立链接
	RedisCli = redis.NewClient(&redis.Options{
		Addr:net.JoinHostPort(host, port),
		DB:db,
		Password:password,
	})
	//测试连接
	if err := RedisCli.Ping().Err(); err != nil {
		panic(err)
	}
}

//RedisCache 缓存查询
func RedisCache(cacheKey string, d interface{}, ex time.Duration, f func() (interface{})) {
	//存在缓存查询缓存
	if v := RedisCli.Get(cacheKey); v != nil && v.Val() != "" {
		if err := json.Unmarshal([]byte(v.Val()), d); err != nil {
			panic(err)
		}
		Logger.Debug("从缓存中读取数据成功", zap.String("cacheKey", cacheKey), zap.String("raw",v.Val()))
		return
	}
	//不存在缓存查询源数据,并保存缓存
	dd := f()
	if dd == nil {
		return
	}

	//保存缓存数据
	dv := reflect.ValueOf(d)
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		panic("the data not ptr or is nil")
	}
	if ddv := reflect.ValueOf(dd); ddv.Kind() == reflect.Ptr {
		dv.Elem().Set(ddv.Elem())
	} else {
		dv.Elem().Set(ddv)
	}

	if b, err := json.Marshal(d); err != nil {
		panic(err)
	} else {
		RedisCli.Set(cacheKey, b, ex)
	}
	Logger.Debug("写入缓存数据", zap.String("cacheKey", cacheKey), zap.String("duration", ex.String()))
}

//ExpireKeys 给永久键设置过期时间
func ExpireKeys(duration time.Duration, keys ...string) {
	for _, key := range keys {
		if RedisCli.TTL(key).Val() == -time.Second {
			if err := RedisCli.Expire(key, duration).Err(); err != nil {
				panic(err)
			}
		}
	}
}

//SyncAction redis防重操作拦截
// 与OnceAction的区别是：onceAction在拦截后即使执行结束，后续的相同操作也不能操作，
// SyncAction用于保存同一时间只能有一个操作
func SyncAction(key string, expiration time.Duration, callback func()) error {
	if !RedisCli.SetNX(key, 1, expiration).Val() {
		//有其他线程正在操作，本次返回
		return fmt.Errorf(key+" RedisSyncAction已被其他协程占用，本次执行结束")
	}
	defer RedisCli.Del(key)
	callback()
	return nil
}

//OnceAction 一段时间内幂等拦截，效果不同于mysql的永久幂等
// 利用redis做一次操作拦截,且在一段时间内都不允许重复,仅在panic下不拦截一次操作
func OnceAction(traceId string, actionName string, duration time.Duration, fun func()) {
	//查看是否已经操作过
	onceKey := fmt.Sprintf("once_action:action:%s:trace_id:%s", actionName, traceId)
	if !RedisCli.SetNX(onceKey, 1, duration).Val() {
		return
	}

	//如果抛出错误则删除操作一次拦截
	defer func() {
		if e := recover(); e != nil {
			RedisCli.Del(onceKey)
			panic(e)
		}
	}()

	//执行操作
	fun()
}

//RedisListScan redis list集合扫描,
// fun(start int64, data []string)
func RedisListScan(listKey string, limit int64, fun func(int64, []string)) {
	start := int64(0)
	for {
		//获取投票的用户id
		data, err := RedisCli.LRange(listKey, start*limit, (start+1)*limit-1).Result()
		if err != nil && err != redis.Nil {
			panic(err)
		}

		//没有更多数据，执行结束
		if len(data) == 0 {
			break
		}

		//执行处理逻辑
		fun(start*limit, data)

		start++
	}
}