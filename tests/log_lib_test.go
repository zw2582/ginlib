package tests

import (
	"github.com/zw2582/ginlib"
	"os"
	"testing"
	"time"
)

//测试根据信号切割日志
func TestRotateWithSign(t *testing.T)  {
	ginlib.InitIni("./conf/app.ini")
	ginlib.InitLogger()

	t.Log("pid", os.Getpid())

	for i := 0; i< 1000; i++ {
		ginlib.Logger.Info("你哈哈哈哈哈哈哈哈哈哈哈哈")
		time.Sleep(time.Second)
	}
}
