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

var (
	DB *gorm.DB
)

//InitDb 初始化数据库
func InitDb() *gorm.DB {
	host := Ini_Str("mysql.host")
	name := Ini_Str("mysql.name")
	password := Ini_Str("mysql.password")
	database := Ini_Str("mysql.database")
	port := Ini_Int("mysql.port", 3306)
	conn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&interpolateParams=True", name, password, host, port, database)

	eng, err := gorm.Open(mysql.Open(conn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: logger.New(GormLogger{}, logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,  // Ignore ErrRecordNotFound error for logger
			Colorful:                  false, // Disable color
		}),
	})
	if err != nil {
		panic(err)
	}
	DB = eng

	return DB
}

//GormLogger 定义gorm日志
type GormLogger struct{}

func (this GormLogger) Printf(format string, v ...interface{}) {
	//设置日志
	showSql := Ini_Bool("mysql.show_sql")
	if !showSql {
		return
	}

	format = strings.ReplaceAll(format, "\n", " ")
	path := v[0]
	v[0] = "MYSQL"
	Logger.WithOptions(zap.AddCallerSkip(4)).Info(fmt.Sprintf(format, v...), zap.Any("file_path", path))
}
