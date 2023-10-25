package ginlib

import (
	"ad-dashboard-statistics-auth/utils"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"strings"
	"time"
)

var (
	//BackMysqlDB 看板mysql
	BackMysqlDB *gorm.DB

	//MongoDBKdj 看短剧MongoDB
	MongoDBKdj *mongo.Database
)

// InitBackMysqlDB 看板数据库初始化
func InitBackMysqlDB() {
	dsn := utils.Ini_Str("mysql_back.dsn")
	maxOpenConn := utils.Ini_Int("mysql_back.max_open_conn")
	maxIdleConn := utils.Ini_Int("mysql_back.max_idle_conn")
	maxLifeSecond := utils.Ini_Int("mysql_back.max_life_second")
	showSqlLog := utils.Ini_Bool("mysql_back.show_sql")
	BackMysqlDB = initDB(dsn, maxOpenConn, maxIdleConn, maxLifeSecond, showSqlLog)
}

// InitMongoDBKdj 初始化看短剧mongo
func InitMongoDBKdj() {
	uri := utils.Ini_Str("mongo_kdj.uri")
	database := utils.Ini_Str("mongo_kdj.database")
	maxPoolSize := utils.Ini_Int("mongo_kdj.max_pool_size")
	maxIdleSecond := utils.Ini_Int("mongo_kdj.max_idle_second")
	MongoDBKdj = initMongoDB(uri, database, maxPoolSize, maxIdleSecond)
}

// InitDb 初始化数据库
func initDB(dsn string, maxOpenConn, maxIdleConn, maxLifeSecond int, showSqlLog bool) *gorm.DB {
	log.Printf("初始化MysqlDB dsn:%s maxOpenConn:%d maxIdleConn:%d maxLifeSecond:%ds \n", dsn, maxOpenConn, maxIdleConn, maxLifeSecond)
	c := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}
	if showSqlLog {
		c.Logger = logger.New(gormLogger{}, logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,  // Ignore ErrRecordNotFound error for logger
			Colorful:                  false, // Disable color
		})
	}
	eng, err := gorm.Open(mysql.Open(dsn), c)
	if err != nil {
		panic(err)
	}
	sqlDb, _ := eng.DB()
	sqlDb.SetMaxOpenConns(maxOpenConn)
	sqlDb.SetMaxIdleConns(maxIdleConn)
	sqlDb.SetConnMaxLifetime(time.Second * time.Duration(maxLifeSecond))
	return eng
}

// gormLogger 定义gorm日志
type gormLogger struct {
}

func (this gormLogger) Printf(format string, v ...interface{}) {
	format = strings.ReplaceAll(format, "\n", " ")
	v[0] = "[MysqlDB]"
	utils.Logger.WithOptions(zap.AddCallerSkip(4)).Info(fmt.Sprintf(format, v...))
}

// initMongoDB 初始化mongodb
func initMongoDB(uri, database string, maxPoolSize, maxIdleTime int) *mongo.Database {
	o := options.Client()
	o.ApplyURI(uri)
	o.SetMaxPoolSize(uint64(maxPoolSize))
	o.SetMaxConnIdleTime(time.Duration(maxIdleTime) * time.Second)
	o.SetLoggerOptions(options.Logger().SetSink(mongoLogger{}).SetComponentLevel(options.LogComponentCommand, options.LogLevelInfo))
	log.Printf("初始化MongoDB uri:%s database:%s maxPoolSize:%d maxIdleTime:%ds \n", uri, database, maxPoolSize, maxIdleTime)
	client, err := mongo.Connect(context.TODO(), o)
	if err != nil {
		panic(err)
	}
	// 检查连接
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	return client.Database(database)
}

type mongoLogger struct {
}

func (m mongoLogger) Info(level int, message string, keysAndValues ...interface{}) {
	params := make([]zap.Field, 0)
	l := len(keysAndValues)
	for i := 0; i < l; i += 2 {
		params = append(params, zap.Any(fmt.Sprintf("%v", keysAndValues[i]), keysAndValues[i+1]))
	}
	utils.Logger.WithOptions(zap.AddCallerSkip(6)).Info("[MongoDB] "+message, params...)
}

func (m mongoLogger) Error(err error, message string, keysAndValues ...interface{}) {
	params := make([]zap.Field, 0)
	params = append(params, zap.Error(err))
	l := len(keysAndValues)
	for i := 0; i < l; i += 2 {
		params = append(params, zap.Any(fmt.Sprintf("%v", keysAndValues[i]), keysAndValues[i+1]))
	}
	utils.Logger.WithOptions(zap.AddCallerSkip(6)).Error("[MongoDB] "+message, params...)
}
