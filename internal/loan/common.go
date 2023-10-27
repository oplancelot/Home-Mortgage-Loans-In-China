// parseDate 解析日期字符串并返回时间。如果出现错误，将返回一个零值时间。
package loan

import "time"

func parseDate(dateString string) time.Time {
	layout := "2006-01-02" // 统一的日期布局字符串
	parsedTime, err := time.Parse(layout, dateString)
	if err != nil {
		// 返回零值时间
		return time.Time{}
	}
	return parsedTime
}
