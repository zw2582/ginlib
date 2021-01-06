package conf_demo

import (
	"ginlib"
)

func main() {
	//初始化ini配置
	ginlib.InitIni()

	//初始化日志
	ginlib.InitLogger()

	//初始化redis
	ginlib.InitRedis()

}
