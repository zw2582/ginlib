package ginlib

import (
	"github.com/xormplus/xorm"
	"log"
)

var (
	//待同步校验的表对象
	sync_objs = make([]interface{}, 0)
)

//增加待同步的对象
func XormAddWaitSync(obj interface{})  {
	sync_objs = append(sync_objs, obj)
}

//开始同步表结构
func XormSync(engine *xorm.Engine) (err error) {
	log.Println("同步表结构")
	err = engine.Sync2(sync_objs...)
	if err != nil {
		log.Printf("同步表结构失败:%s\n", err.Error())
	}
	return
}
