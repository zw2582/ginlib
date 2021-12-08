package ginlib

import "time"

//DateSunDayFmt 查询本周周日是几号
func DateSunDayFmt(t time.Time) string {
	return DateSunDay(t).Format("2006-01-02")
}

func DateSunDay(t time.Time) time.Time {
	day := t.Weekday()
	if day == 0 {
		day = 7
		day = 7
	}
	//本周周日是几号
	endDay := t.Add(time.Hour * 24 * time.Duration(7-day))
	return endDay
}
