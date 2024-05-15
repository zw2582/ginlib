package ginlib

import "time"

// LocalUtc7 获取utc+7时区
func LocalUtc7() *time.Location {
	//获取越南时间
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		panic(err)
	}
	return loc
}

// LocalUtc3 获取utc+3时区
func LocalUtc3() *time.Location {
	//获取莫斯科时间
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic(err)
	}
	return loc
}
