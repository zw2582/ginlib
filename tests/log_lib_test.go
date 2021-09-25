package tests

import (
	"github.com/zw2582/ginlib"
	"os"
	"syscall"
	"testing"
	"time"
)

//测试日志
func TestLogger(t *testing.T)  {
	ginlib.InitIni("./conf/app.ini")
	ginlib.InitLogger()

	t.Log("pid", os.Getpid())

	for i := 0; i< 10; i++ {
		ginlib.Logger.Info("你哈哈哈哈哈哈哈哈哈哈哈哈")
		time.Sleep(time.Second)
	}
}

//测试根据信号切割日志
func TestRotateWithSign(t *testing.T)  {
	ginlib.InitIni("./conf/app.ini")
	ginlib.InitLogger(syscall.SIGINT)

	t.Log("pid", os.Getpid())

	for i := 0; i< 100; i++ {
		ginlib.Logger.Info("你哈哈哈哈哈哈哈哈哈哈哈哈")
		time.Sleep(time.Second)
	}
}
