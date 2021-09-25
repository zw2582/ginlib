package tests

import (
	"github.com/zw2582/ginlib"
	"strings"
	"testing"
	"time"
)

func TestOrderNo(t *testing.T)  {
	v := ginlib.OrderNo("a")
	t.Log(v)
}

func TestInvicateCode(t *testing.T)  {
	v,err := ginlib.InviteCode(600)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(v)
	}
}

func TestRedisCache(t *testing.T)  {
	ginlib.InitIni("d:/workspace/ginlib/conf_demo/app.ini")
	ginlib.InitLogger()
	ginlib.InitRedis()
	var a int
	ginlib.RedisCache("caca_test3", &a, time.Minute, func() interface{} {
		return 55
	})

	t.Log("a:", a)
}

func TestSplit(t *testing.T)  {
	a := ""

	aa := strings.Split(a, ",")

	t.Log(len(aa), aa)
}