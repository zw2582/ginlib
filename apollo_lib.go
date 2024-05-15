package ginlib

import (
	"fmt"
	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/storage"
	"go.uber.org/zap"
	"os"
	"strings"
)

var (
	apolloConfigProject *storage.Config
	apolloConfigDefault *storage.Config
	ProjectName         string
)

// Init 阿波罗客户端
func Init(project string, opt ...ApoOption) {
	//loggerDefaultInit(project)
	ProjectName = project

	//定义默认参数
	var defaultIp string
	switch GetEnv() {
	case "local":
		defaultIp = "http://106.54.227.205:8080"
	case "test":
		defaultIp = "http://106.54.227.205:8080"
	case "online":
		defaultIp = "http://106.54.227.205:8080"
	}
	defaultAppId := "karawan"
	defaultSecret := "sfsdfsfdsf"

	//初始化项目配置
	apolloConfigProject = apolloStorageCreate(project, defaultIp, defaultAppId, defaultSecret, opt...)
	//初始化默认配置
	apolloConfigDefault = apolloStorageCreate("default", defaultIp, defaultAppId, defaultSecret, opt...)
	Logger.Info("初始化Apollo配置成功")

	//初始化日志
}

func apolloStorageCreate(cluster, defaultIp, defaultAppId, defaultSecret string, opt ...ApoOption) *storage.Config {
	c := &config.AppConfig{
		AppID:            defaultAppId,
		Cluster:          cluster,
		NamespaceName:    "application.yml",
		IP:               defaultIp,
		IsBackupConfig:   true,
		BackupConfigPath: ".apollo_backup",
		Secret:           defaultSecret,
		MustStart:        true,
	}
	for _, o := range opt {
		o(c)
	}
	Logger.Info("初始化Apollo配置", zap.Any("config", *c))
	var err error
	agollo.SetLogger(&apolloLogger{})
	client, err := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return c, nil
	})
	if err != nil {
		panic(err)
	}
	sc := client.GetConfig("application.yml")
	if sc == nil {
		panic("在阿波罗中未找到application.yml配置文件")
	}
	client.AddChangeListener(apolloChangeLister{cluster: cluster})
	return sc
}

type apolloLogger struct {
}

func (a apolloLogger) Debugf(format string, params ...interface{}) {
	Logger.Debug(fmt.Sprintf("[apollo] "+format, params...))
}

func (a apolloLogger) Infof(format string, params ...interface{}) {
	Logger.Info(fmt.Sprintf("[apollo] "+format, params...))
}

func (a apolloLogger) Warnf(format string, params ...interface{}) {
	Logger.Warn(fmt.Sprintf("[apollo] "+format, params...))
}

func (a apolloLogger) Errorf(format string, params ...interface{}) {
	if strings.Index(format, "get config value fail") > -1 {
		return
	}
	Logger.Error(fmt.Sprintf("[apollo] "+format, params...))
}

func (a apolloLogger) Debug(v ...interface{}) {
	fields := make([]zap.Field, 0)
	for idx, val := range v {
		fields = append(fields, zap.Any(fmt.Sprintf("p%d", idx), val))
	}
	Logger.Debug("[apollo]", fields...)
}

func (a apolloLogger) Info(v ...interface{}) {
	fields := make([]zap.Field, 0)
	for idx, val := range v {
		fields = append(fields, zap.Any(fmt.Sprintf("p%d", idx), val))
	}
	Logger.Info("[apollo]", fields...)
}

func (a apolloLogger) Warn(v ...interface{}) {
	fields := make([]zap.Field, 0)
	for idx, val := range v {
		fields = append(fields, zap.Any(fmt.Sprintf("p%d", idx), val))
	}
	Logger.Warn("[apollo]", fields...)
}

func (a apolloLogger) Error(v ...interface{}) {
	fields := make([]zap.Field, 0)
	for idx, val := range v {
		fields = append(fields, zap.Any(fmt.Sprintf("p%d", idx), val))
	}
	Logger.Error("[apollo]", fields...)
}

// GetEnv 当前环境
func GetEnv() string {
	env := os.Getenv("ENVIRON")
	if env == "" {
		return "local"
	}
	return env
}

// ServerConfigGet 获取内网服务请求地址
func ServerConfigGet(name string) (host, caller, secret string) {
	host = ConfigVal(fmt.Sprintf("server.%s.host", name))
	caller = ConfigVal(fmt.Sprintf("server.%s.caller", name))
	secret = ConfigVal(fmt.Sprintf("server.%s.secret", name))
	return
}

// MysqlConfigGet 获取mysql配置信息
func MysqlConfigGet(name string) (host, user, pwd, db string) {
	host = ConfigVal(fmt.Sprintf("mysql.%s.host", name))
	user = ConfigVal(fmt.Sprintf("mysql.%s.user", name))
	pwd = ConfigVal(fmt.Sprintf("mysql.%s.pwd", name))
	db = ConfigVal(fmt.Sprintf("mysql.%s.db", name))
	return
}

// RedisConfigGet 获取redis的配置信息
func RedisConfigGet(name string) (host, pwd string, db, poolSize int) {
	host = ConfigVal(fmt.Sprintf("redis.%s.host", name))
	pwd = ConfigVal(fmt.Sprintf("redis.%s.pwd", name))
	db = ConfigInt(fmt.Sprintf("redis.%s.db", name), 0)
	poolSize = ConfigInt(fmt.Sprintf("redis.%s.pool_size", name), 10)
	return
}

func ConfigVal(key string) string {
	tmp := apolloConfigProject.GetValue(key)
	if tmp == "" {
		tmp = apolloConfigDefault.GetValue(key)
	}
	return tmp
}

func ConfigStr(key, defaultValue string) string {
	return apolloConfigProject.GetStringValue(key, apolloConfigDefault.GetStringValue(key, defaultValue))
}

func ConfigInt(key string, defaultValue int) int {
	return apolloConfigProject.GetIntValue(key, apolloConfigDefault.GetIntValue(key, defaultValue))
}

func ConfigFloat(key string, defaultValue float64) float64 {
	return apolloConfigProject.GetFloatValue(key, apolloConfigDefault.GetFloatValue(key, defaultValue))
}

func ConfigBool(key string, defaultValue bool) bool {
	return apolloConfigProject.GetBoolValue(key, apolloConfigDefault.GetBoolValue(key, defaultValue))
}

func ConfigStrSlice(key string, defaultValue []string, separator ...string) []string {
	sep := ","
	if len(separator) > 0 {
		sep = separator[0]
	}
	return apolloConfigProject.GetStringSliceValue(key, sep, apolloConfigDefault.GetStringSliceValue(key, sep, defaultValue))
}

func ConfigIntSlice(key string, defaultValue []int, separator ...string) []int {
	sep := ","
	if len(separator) > 0 {
		sep = separator[0]
	}
	return apolloConfigProject.GetIntSliceValue(key, sep, apolloConfigDefault.GetIntSliceValue(key, sep, defaultValue))
}

type ApoOption func(c *config.AppConfig)

// WithApolloIp 配置阿波罗ip
func WithApolloIp(ip string) ApoOption {
	return func(c *config.AppConfig) {
		c.IP = ip
	}
}

// WithApolloAppId 配置阿波罗appId
func WithApolloAppId(appId string) ApoOption {
	return func(c *config.AppConfig) {
		c.AppID = appId
	}
}

func WithApolloSecret(secret string) ApoOption {
	return func(c *config.AppConfig) {
		c.Secret = secret
	}
}

type apolloChangeLister struct {
	cluster string
}

func (a apolloChangeLister) OnChange(event *storage.ChangeEvent) {
	Logger.Info("[apollo] OnChange", zap.String("Cluster", a.cluster), zap.Any("changes", event.Changes), zap.Int64("NotificationID", event.NotificationID), zap.String("Namespace", event.Namespace))
	//检测日志变化
	logChanged := false
	logLabs := []string{"log.path", "log.level", "log.encode", "log.sql"}
	for _, val := range logLabs {
		if event.Changes[val] != nil {
			logChanged = true
			break
		}
	}
	if logChanged {
		Logger.Info("重启logger")
	}
}

func (a apolloChangeLister) OnNewestChange(event *storage.FullChangeEvent) {

}
