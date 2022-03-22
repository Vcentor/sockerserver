// @Author: Vcentor
// @Date: 2020/11/13 1:07 下午

package utils

import (
	"time"
)

// TimeStamp2Datetime 时间戳转换成日期时间
func TimeStamp2Datetime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

//GetNowDate 获取当前日期
func GetNowDate(split string) string {
	y := time.Now().Format("2006")
	m := time.Now().Format("01")
	d := time.Now().Format("02")
	return y + split + m + split + d
}

// Datetime2Timestamp 日期时间转时间戳
func Datetime2Timestamp(datetime string) (int64, error) {
	var timestamp int64
	tmp := "2006-01-02 15:04:05"
	t, err := time.ParseInLocation(tmp, datetime, time.Local)
	if err != nil {
		return timestamp, err
	}
	return t.Unix(), nil
}
