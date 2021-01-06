package conf_demo

import (
	"weishen_gin_lib"
)

func main() {
	//初始化ini配置
	weishen_gin_lib.InitIni()

	//初始化日志
	weishen_gin_lib.InitLogger()

	//初始化redis
	weishen_gin_lib.InitRedis()

}
