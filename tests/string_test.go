package tests

import (
	"testing"
	"ginlib"
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