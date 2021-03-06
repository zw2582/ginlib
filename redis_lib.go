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

//RedisSyncAction redis实现同步操作，解决并发问题
func RedisSyncAction(key string, expiration time.Duration, callback func()) error {
	if !RedisCli.SetNX(key, 1, expiration).Val() {
		//有其他线程正在操作，本次返回
		return fmt.Errorf(key+" RedisSyncAction已被其他协程占用，本次执行结束")
	}
	defer RedisCli.Del(key)
	callback()
	return nil
}