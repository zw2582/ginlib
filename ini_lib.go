package ginlib

import (
	"fmt"
	"github.com/go-ini/ini"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	iniFile    *ini.File
	APP_NAME   string
	APP_HOST   string
	APP_PORT   string
	envPattern = regexp.MustCompile(`\$\{([^)}{(]+)}$`)
)

// InitIni 初始化加载配置文件
// 支持环境变量读取 使用${xxx}，读取环境变量为xxx的值
func InitIni(inipath ...string) {
	defaultPath := "conf/app.dev.ini"
	if len(inipath) > 0 {
		defaultPath = inipath[0]
	}
	log.Println("初始化加载配置文件", inipath)
	t, err := ini.Load(defaultPath)
	if err != nil {
		panic(err)
	}
	iniFile = t
	//设置gin运行环境
	APP_NAME = iniFile.Section("app").Key("name").Value()
	APP_HOST = iniFile.Section("app").Key("host").Value()
	APP_PORT = iniFile.Section("app").Key("port").Value()
}

// Ini_Str 读取配置文件信息 key格式可以是“section.key”
func Ini_Str(key string, defaults ...string) string {
	value, exist := IniValueFetch(key)
	if !exist {
		if len(defaults) > 0 {
			return defaults[0]
		}
		return ""
	}
	return value
}

// Ini_Int 获取默认int值
func Ini_Int(key string, defaults ...int) int {
	def := 0
	if len(defaults) > 0 {
		def = defaults[0]
	}
	value, exist := IniValueFetch(key)
	if !exist {
		return def
	}
	val, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return val
}

// Ini_Bool 获取默认Bool值
func Ini_Bool(key string, defaults ...bool) bool {
	def := false
	if len(defaults) > 0 {
		def = defaults[0]
	}
	value, exist := IniValueFetch(key)
	if !exist {
		return def
	}
	val, err := parseBool(value)
	if err != nil {
		return def
	}
	return val
}

// GetEnv 获取环境变量:ENVIRON
func GetEnv() string {
	env := os.Getenv("ENVIRON")
	if env == "" {
		return "dev"
	}
	return env
}

// IniValueFetch 获取init的数据，支持key自动切分，支持环境变量读取
// exist的判断主要是根据ini文件中是否存在key判断，而不是根据value是否为空字符串判断，但是如果需要查询os的环境变量，就会根据value判断是否存在
func IniValueFetch(key string) (value string, exist bool) {
	keys := strings.Split(key, ".")
	section := ""
	if len(keys) == 2 {
		section = keys[0]
		key = keys[1]
	}
	if !iniFile.Section(section).HasKey(key) {
		return "", false
	}
	exist = true
	value = iniFile.Section(section).Key(key).Validate(func(s string) string {
		//使其支持环境变量
		vr := envPattern.FindString(s)
		if len(vr) == 0 {
			return s
		}
		osKey := vr[2 : len(vr)-1]
		osVal := os.Getenv(osKey)
		if osVal == "" {
			exist = false
		}
		return osVal
	})
	return value, exist
}

func parseBool(str string) (value bool, err error) {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True", "YES", "yes", "Yes", "y", "ON", "on", "On":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False", "NO", "no", "No", "n", "OFF", "off", "Off":
		return false, nil
	}
	return false, fmt.Errorf("parsing \"%s\": invalid syntax", str)
}
