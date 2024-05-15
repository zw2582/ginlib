package ginlib

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"strings"
	"time"
)

// MysqlDBCreate 创建mysql连接
func MysqlDBCreate(dsn string, maxOpenConn, maxIdleConn, maxLifeSecond int, showLog bool) *gorm.DB {
	eng, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: logger.New(GormLogger{showLog}, logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,  // Ignore ErrRecordNotFound error for logger
			Colorful:                  false, // Disable color
		}),
	})
	if err != nil {
		panic(err)
	}
	sqlDb, _ := eng.DB()
	sqlDb.SetMaxOpenConns(maxOpenConn)
	sqlDb.SetMaxIdleConns(maxIdleConn)
	sqlDb.SetConnMaxLifetime(time.Second * time.Duration(maxLifeSecond))

	return eng
}

// GormLogger 定义gorm日志
type GormLogger struct {
	ShowLog bool
}

func (this GormLogger) Printf(format string, v ...interface{}) {
	if !this.ShowLog {
		return
	}
	format = strings.ReplaceAll(format, "\n", " ")
	v[0] = "MYSQL"
	Logger.WithOptions(zap.AddCallerSkip(4)).Info(fmt.Sprintf(format, v...))
}
