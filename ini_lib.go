package ginlib

import (
	"github.com/go-ini/ini"
	"log"
	"strings"
)

var (
	iniFile *ini.File
	APP_NAME string
	APP_HOST string
	APP_PORT string
	APP_ENV string
)

//InitIni 初始化加载配置文件
func InitIni(inipath ...string)  {
	defaultPath := "conf/app.ini"
	if len(inipath) > 0 {
		defaultPath = inipath[0]
	}
	log.Println("初始化加载配置文件")
	t, err := ini.Load(defaultPath)
	if err != nil {
		panic(err)
	}
	iniFile = t
	//设置gin运行环境
	APP_NAME = iniFile.Section("app").Key("name").Value()
	APP_HOST = iniFile.Section("app").Key("host").Value()
	APP_PORT = iniFile.Section("app").Key("port").Value()
	APP_ENV = iniFile.Section("app").Key("env").Value()
}

//Ini_Str 读取配置文件信息 key格式可以是“section.key”
func Ini_Str(key string, defaults ...string) string {
	keys := strings.Split(key, ".")
	section := ""
	if len(keys) == 2 {
		section = keys[0]
		key = keys[1]
	}
	if !iniFile.Section(section).HasKey(key) {
		if len(defaults) > 0 {
			return defaults[0]
		}
		return ""
	}
	value := iniFile.Section(section).Key(key).Value()
	return value
}

//Ini_Int 获取默认int值
func Ini_Int(key string, defaults ...int) int {
	def:= 0
	if len(defaults) > 0 {
		def = defaults[0]
	}
	keys := strings.Split(key, ".")
	section := ""
	if len(keys) == 2 {
		section = keys[0]
		key = keys[1]
	}
	if !iniFile.Section(section).HasKey(key) {
		return def
	}
	value, err := iniFile.Section(section).Key(key).Int()
	if err != nil {
		return def
	}
	return value
}

//Ini_Bool 获取默认Bool值
func Ini_Bool(key string, defaults ...bool) bool {
	def:= false
	if len(defaults) > 0 {
		def = defaults[0]
	}
	keys := strings.Split(key, ".")
	section := ""
	if len(keys) == 2 {
		section = keys[0]
		key = keys[1]
	}
	if !iniFile.Section(section).HasKey(key) {
		return def
	}
	value, err := iniFile.Section(section).Key(key).Bool()
	if err != nil {
		return def
	}
	return value
}