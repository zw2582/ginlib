package tests

import (
	"testing"
	"weishen_gin_lib"
)

func TestOrderNo(t *testing.T)  {
	v := weishen_gin_lib.OrderNo("a")
	t.Log(v)
}

func TestInvicateCode(t *testing.T)  {
	v,err := weishen_gin_lib.InviteCode(600)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(v)
	}
}