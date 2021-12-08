package ginlib

import (
	"github.com/jinzhu/gorm"
	"log"
)

var (
	//待同步校验的表对象
	sync_objs = make([]interface{}, 0)
)

//GormAddWaitSync 增加待同步的对象
func GormAddWaitSync(obj interface{})  {
	sync_objs = append(sync_objs, obj)
}

//GormSync 开始同步表结构
func GormSync(db *gorm.DB) (err error) {
	log.Println("同步表结构")

	err = db.AutoMigrate(sync_objs...).Error
	if err != nil {
		log.Printf("同步表结构失败:%s\n", err.Error())
	}
	return
}

//OrmGetSyncObjs 获取需要同步的数据库对象
func OrmGetSyncObjs() []interface{} {
	return sync_objs
}