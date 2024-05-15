package utils

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"go.uber.org/zap"
	"runtime"
)

var (
	//MetricCntFail 错误请求数
	MetricCntFail = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongo_cnt_fail",
	}, []string{"db", "path"})

	//MetricDuration 请求耗时
	MetricDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "mongo_duration",
		Buckets: []float64{1, 3, 5, 7, 9, 15, 25, 50},
	}, []string{"db", "path"})
)

// MongoDBCreate 初始化mongo数据库
func MongoDBCreate(ctx context.Context, connUri string, showLog bool) *mongo.Database {
	cs, err := connstring.Parse(connUri)
	if err != nil {
		panic(err)
	}
	if cs.Database == "" {
		panic("请指定默认MongoDB数据库")
	}
	clientOps := options.Client().ApplyURI(connUri)

	//定义日志
	if showLog {
		clientOps = clientOps.SetLoggerOptions(&options.LoggerOptions{
			ComponentLevels: map[options.LogComponent]options.LogLevel{
				options.LogComponentCommand: options.LogLevelDebug,
			},
			Sink:              MongoSink{},
			MaxDocumentLength: 2048,
		})
	}

	//增加监控
	clientOps = clientOps.SetMonitor(&event.CommandMonitor{
		Started: nil,
		Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
			if evt.CommandName == "ping" {
				return
			}
			_, file, line, ok := runtime.Caller(6)
			if ok {
				path := fmt.Sprintf("%s:%d", file, line)
				MetricDuration.WithLabelValues(evt.DatabaseName, path).Observe(float64(evt.Duration.Milliseconds()))
			}
		},
		Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
			if evt.CommandName == "ping" {
				return
			}
			_, file, line, ok := runtime.Caller(6)
			if ok {
				path := fmt.Sprintf("%s:%d", file, line)
				MetricCntFail.WithLabelValues(evt.DatabaseName, path).Inc()
			}
		},
	})

	client, err := mongo.Connect(ctx, clientOps)
	if err != nil {
		panic(err)
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		panic("MongoDB链接失败")
	}

	mongoDb := client.Database(cs.Database)
	return mongoDb
}

type MongoSink struct {
}

func (m MongoSink) Info(level int, message string, keysAndValues ...interface{}) {
	fields := []zap.Field{zap.Int("level", level)}
	for i := 1; i < len(keysAndValues); i += 2 {
		fields = append(fields, zap.Any(fmt.Sprintf("%v", keysAndValues[i-1]), keysAndValues[i]))
	}
	Logger.WithOptions(zap.AddCallerSkip(7)).Info("[Mongo]"+message, fields...)
}

func (m MongoSink) Error(err error, message string, keysAndValues ...interface{}) {
	fields := make([]zap.Field, 0)
	fields = append(fields, zap.Error(err))
	for i := 1; i < len(keysAndValues); i += 2 {
		fields = append(fields, zap.Any(fmt.Sprintf("%v", keysAndValues[i-1]), keysAndValues[i]))
	}
	Logger.WithOptions(zap.AddCallerSkip(7)).Error("[Mongo]"+message, fields...)
}
