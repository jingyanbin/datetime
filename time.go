package datetime

import (
	. "github.com/jingyanbin/base"
	. "github.com/jingyanbin/timezone"
	"strconv"
	_ "unsafe"
)

const formatterYmdHMS = "%Y/%m/%d %H:%M:%S"
const minSec = 60
const hourSec = 3600
const daySec = 3600 * 24 //每天的秒数
const weekSec = 3600 * 24 * 7

const firstYears = 365
const secondYears = 365 + 365
const thirdYears = 365 + 365 + 366
const fourYears = 365 + 365 + 366 + 365 //每个四年的总天数

var norMonth = [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}  //平年
var leapMonth = [12]int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31} //闰年

//是否是闰年
func leapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || year%400 == 0
}

func checkClock(hour, min, sec int) error {
	if hour > 23 || hour < 0 {
		return NewError("check date clock error: out of range hour=%v", hour)
	}
	if min < 0 || min > 59 {
		return NewError("check date clock error: out of range min=%v", min)
	}
	if sec < 0 || sec > 59 {
		return NewError("check date clock error: out of range sec=%v", sec)
	}
	return nil
}

func checkDateClock(year, month, day, hour, min, sec int) error {
	if year > 9999 || year < 1 {
		return NewError("check date clock error: out of range year=%v", year)
	}
	if month > 12 || month < 1 {
		return NewError("check date clock error: out of range month=%v", month)
	}
	if day < 1 || day > 31 {
		return NewError("check date clock error: out of range day=%v", day)
	}

	m := month % 7
	if m == 0 {
		m = 1
	}
	if month == 2 {
		if leapYear(year) {
			if day > 29 {
				return NewError("check date clock error: out of range day=%v", day)
			}
		} else {
			if day > 28 {
				return NewError("check date clock error: out of range day=%v", day)
			}
		}
	} else {
		if m%2 == 1 {
			if day > 31 {
				return NewError("check date clock error: out of range day=%v", day)
			}
		} else {
			if day > 30 {
				return NewError("check date clock error: out of range day=%v", day)
			}
		}
	}
	return checkClock(hour, min, sec)
}

//@description: 年,月,日,时,分,秒 -> 日期时间字符串
//@param:       year, month, day, hour, min, sec "年,月,日,时,分,秒"
//@param:       formatter string "格式化模板" 如: "%Y/%m/%d %H:%M:%S", "%Y-%m-%d %H:%M:%S", "%Y%m%d%H%M%S"
//@param:       string "日期时间字符串"
func DateClockToFormat(year, month, day, hour, min, sec int, formatter string) string {
	var theTime []byte
	length := len(formatter)
	for i := 0; i < length; {
		c := formatter[i]
		if c == '%' {
			if i+1 == length {
				break
			}
			c2 := formatter[i+1]
			switch c2 {
			case 'Y': //四位数的年份表示（0000-9999）
				ItoAW(&theTime, year, 4)
			case 'm': //月份（01-12）
				ItoAW(&theTime, month, 2)
			case 'd': //月内中的一天（0-31）
				ItoAW(&theTime, day, 2)
			case 'H': //24小时制小时数（0-23）
				ItoAW(&theTime, hour, 2)
			case 'M': //分钟数（00=59）
				ItoAW(&theTime, min, 2)
			case 'S': //秒（00-59）
				ItoAW(&theTime, sec, 2)
			default:
				theTime = append(theTime, c2)
			}
			i += 2
		} else {
			theTime = append(theTime, c)
			i += 1
		}
	}
	return string(theTime)
}

//@description: 年,月,日,时,分,秒 -> 换为标准日期时间字符串
//@param:       year, month, day, hour, min, sec "年,月,日,时,分,秒"
//@param:       string "日期时间字符串"
func DateClockToYmdHMS(year, month, day, hour, min, sec int) string {
	return DateClockToFormat(year, month, day, hour, min, sec, formatterYmdHMS)
}

//@description: 年,月,日,时,分,秒 -> 转换为秒级时间戳
//@param:       year, month, day, hour, min, sec "年,月,日,时,分,秒"
//@param:       zone TimeZone "时区"
//@return:      unix int64 "秒级时间戳"
//@return:      yDay "一年中第几天"
//@return:      daySecond "一天中第几秒"
//@return:      error "错误信息"
func DateClockToUnix(year, month, day, hour, min, sec int, zone TimeZone) (unix int64, yDay int, daySecond int, err error) {
	err = checkDateClock(year, month, day, hour, min, sec)
	if err != nil {
		err = NewError("date clock to unix error: time=%v, err=%v", DateClockToYmdHMS(year, month, day, hour, min, sec), err)
		return
	}

	nCha := year - 1970
	var neg bool
	if nCha < 0 {
		nCha = -nCha
		neg = true
	}

	nYear4 := nCha >> 2
	nYearMod := nCha % 4
	nDays := nYear4 * fourYears
	pMonth := &norMonth
	if nYearMod == 1 {
		nDays += firstYears
	} else if nYearMod == 2 {
		nDays += secondYears
		if !neg {
			pMonth = &leapMonth
		}
	} else if nYearMod == 3 {
		nDays += thirdYears
	}

	if neg {
		nDays = -nDays
	}

	//var yDay int
	for i := 0; i < month-1; i++ {
		nDays += pMonth[i]
		yDay += pMonth[i]
	}

	nDays += day - 1
	daySecond = hour*hourSec + min*minSec + sec
	unix = int64(nDays*daySec+daySecond) - zone.Offset()
	//return unix, yDay, daySecond, nil
	return
}

//@description: 日期时间字符串 -> 根据格式化模板 -> 转换为秒级时间戳
//@param:       s string "日期时间字符串"
//@param:       formatter string "格式化模板" 如: "%Y/%m/%d %H:%M:%S", "%Y-%m-%d %H:%M:%S", "%Y%m%d%H%M%S"
//@param:       zone TimeZone "时区"
//@param:       extend bool "是否启用扩展增强模式" 与函数 FormatToDateClock 一样
//@return:      unix int64 "秒级时间戳"
//@return:      error "错误信息"
func FormatToUnix(s, formatter string, zone TimeZone, extend bool) (unix int64, err error) {
	year, month, day, hour, min, sec, err := FormatToDateClock(s, formatter, extend)
	if err != nil {
		return 0, err
	}
	unix, _, _, err = DateClockToUnix(year, month, day, hour, min, sec, zone)
	if err != nil {
		return 0, err
	}
	return unix, nil
}

//@description: 标准日期时间字符串 -> 转换为秒级时间戳
//@param:       s string "日期时间字符串"
//@param:       zone TimeZone "时区"
//@param:       extend bool "是否启用扩展增强模式" 与函数 FormatToDateClock 一样
//@return:      unix int64 "秒级时间戳"
//@return:      error "错误信息"
func YmdHMSToUnix(s string, zone TimeZone, extend bool) (unix int64, err error) {
	return FormatToUnix(s, formatterYmdHMS, zone, extend)
}

//@description: 秒级时间戳 ->转为 年,月,日,时,分,秒
//@param:       unix int64 "秒级时间戳"
//@param:       zone TimeZone "时区"
//@return:      year, month, day, hour, min, sec "年,月,日,时,分,秒"
//@return:      yDay "一年中第几天"
//@return:      daySecond "一天中第几秒"
func UnixToDateClock(unix int64, zone TimeZone) (year, month, day, hour, min, sec, yDay, daySecond int) {
	unixLocal := unix + zone.Offset()
	var nRemain int
	if unixLocal < 0 {
		nUnixSec := -unixLocal
		nDays := int(nUnixSec / daySec)
		daySecond = (daySec - int(nUnixSec-int64(nDays*daySec))) % daySec
		nYear4 := nDays/fourYears + 1
		nRemain = nYear4*fourYears - nDays
		if daySecond == 0 {
			nRemain += 1
		}
		year = 1970 - nYear4<<2
	} else {
		nDays := int(unixLocal / daySec)
		daySecond = int(unixLocal - int64(nDays*daySec))
		nYear4 := nDays / fourYears
		nRemain = nDays - nYear4*fourYears + 1
		year = 1970 + nYear4<<2
	}
	pMonth := &norMonth
	if nRemain <= firstYears {

	} else if nRemain <= secondYears {
		year += 1
		nRemain -= firstYears
	} else if nRemain <= thirdYears {
		year += 2
		nRemain -= secondYears
		pMonth = &leapMonth
	} else if nRemain <= fourYears {
		year += 3
		nRemain -= thirdYears
	} else {
		year += 4
		nRemain -= fourYears
	}
	yDay = nRemain
	var nTemp int
	for i := 0; i < 12; i++ {
		nTemp = nRemain - pMonth[i]
		if nTemp < 1 {
			month = i + 1
			if nTemp == 0 {
				day = pMonth[i]
			} else {
				day = nRemain
			}
			break
		}
		nRemain = nTemp
	}
	hour = daySecond / hourSec
	inHourSec := daySecond - hour*hourSec
	min = inHourSec / minSec
	sec = inHourSec - min*minSec
	return
}

//@description: 日期时间字符串中获取 -> 年,月,日,时,分,秒
//@param:       s string "日期时间字符串"
//@param:       formatter string "格式化字符串"
//@param:       extend bool "是否启用扩展模式"
//              扩展模式: 可以识别非标准日期时间字符串 如: 2020/1/1 0:1:1
//              非扩展模式: 只能识别标准日期时间字符串 如: 2020/01/01 00:01:01
//@return:      year, month, day, hour, min, sec int "日期时间"
//@return:      error "错误信息"
func FormatToDateClock(s, formatter string, extend bool) (year, month, day, hour, min, sec int, err error) {
	if extend {
		return formatToDateClockEx(s, formatter)
	} else {
		return formatToDateClock(s, formatter)
	}
}

func formatToDateClock(s, formatter string) (year, month, day, hour, min, sec int, err error) {
	defer Exception(func(stack string, e error) {
		err = NewError("format to date clock error: %v, time=%v \n%v", e, s, stack)
	})
	var pos int
	var pos2 int
	length := len(formatter)
	sLen := len(s)
	for i := 0; i < length; {
		c := formatter[i]
		if c == '%' {
			if i+1 == length {
				break
			}
			c2 := formatter[i+1]

			switch c2 {
			case 'Y': //四位数的年份表示（0000-9999）
				pos2 = pos + 4
				if pos2 > sLen {
					err = NewError("format to date clock format length error: to year, time=%v, formatter=%v", s, formatter)
					return
				}
				c3 := s[pos:pos2]
				pos = pos2
				year, err = strconv.Atoi(c3)
				if err != nil {
					err = NewError("format to date clock param error: year=%v, time=%v", c3, s)
					return
				}
			case 'm': //月份（01-12）
				pos2 = pos + 2
				if pos2 > sLen {
					err = NewError("format to date clock format length error: to month, time=%v, formatter=%v", s, formatter)
					return
				}
				c3 := s[pos:pos2]
				pos = pos2
				month, err = strconv.Atoi(c3)
				if err != nil {
					err = NewError("format to date clock param error: month=%v, time=%v", c3, s)
					return
				}
			case 'd': //月内中的一天（01-31）
				pos2 = pos + 2
				if pos2 > sLen {
					err = NewError("format to date clock format length error: to day, time=%v, formatter=%v", s, formatter)
					return
				}
				c3 := s[pos:pos2]
				pos = pos2
				day, err = strconv.Atoi(c3)
				if err != nil {
					err = NewError("format to date clock param error: day=%v, time=%v", c3, s)
					return
				}
			case 'H': //24小时制小时数（00-23）
				pos2 = pos + 2
				if pos2 > sLen {
					err = NewError("format to date clock format length error: to hour, time=%v, formatter=%v", s, formatter)
					return
				}
				c3 := s[pos:pos2]
				pos = pos2
				hour, err = strconv.Atoi(c3)
				if err != nil {
					err = NewError("format to date clock param error: hour=%v, time=%v", c3, s)
					return
				}
			case 'M': //分钟数（00=59）
				pos2 = pos + 2
				if pos2 > sLen {
					err = NewError("format to date clock format length error: to min, time=%v, formatter=%v", s, formatter)
					return
				}
				c3 := s[pos:pos2]
				pos = pos2
				min, err = strconv.Atoi(c3)
				if err != nil {
					err = NewError("format to date clock param error: min=%v, time=%v", c3, s)
					return
				}
			case 'S': //秒（00-59）
				pos2 = pos + 2
				if pos2 > sLen {
					err = NewError("format to date clock format length error: to sec, time=%v, formatter=%v", s, formatter)
				}
				c3 := s[pos:pos2]
				pos = pos2
				sec, err = strconv.Atoi(c3)
				if err != nil {
					err = NewError("format to date clock param error: sec=%v, time=%v", c3, s)
					return
				}
			default:
				err = NewError("format to date clock formatter error: %v, time=%v, formatter=%v", formatter[i:i+2], s, formatter)
				return
			}
			i += 2
		} else {
			if formatter[i] != s[pos] {
				err = NewError("format to date clock separator error: %v, time=%v, formatter=%v", string(s[pos]), s, formatter)
				return
			}
			i += 1
			pos += 1
		}
	}

	err = checkDateClock(year, month, day, hour, min, sec)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, NewError("format to date clock check error: time=%v, err=%v", s, err)
	}
	return
}

func formatToDateClockEx(s, formatter string) (year, month, day, hour, min, sec int, err error) {
	defer Exception(func(stack string, e error) {
		err = NewError("format to date clock ex exception: %v, time=%v \n%v", e, s, stack)
	})
	numbers := NewNextNumber(s)
	var found bool

	length := len(formatter)
	var jump int
	for i := 0; i < length; {
		c := formatter[i]
		if c == '%' {
			if i+1 == length {
				break
			}
			c2 := formatter[i+1]
			switch c2 {
			case 'Y': //四位数的年份表示（0000-9999）
				year, found = numbers.Next(jump, 4)
			case 'm': //月份（01-12）
				month, found = numbers.Next(jump, 2)
			case 'd': //月内中的一天（01-31）
				day, found = numbers.Next(jump, 2)
			case 'H': //24小时制小时数（00-23）
				hour, found = numbers.Next(jump, 2)
			case 'M': //分钟数（00=59）
				min, found = numbers.Next(jump, 2)
			case 'S': //秒（00-59）
				sec, found = numbers.Next(jump, 2)
			default:
				err = NewError("format to date clock ex formatter error: %v, time=%v, formatter=%v", formatter[i:i+2], s, formatter)
				return
			}

			if !found {
				err = NewError("format to date clock ex found number error: time=%v, formatter=%v", s, formatter)
				return
			}
			jump = 0
			i += 2
		} else {
			jump += 1
			i += 1
		}
	}

	err = checkDateClock(year, month, day, hour, min, sec)
	if err != nil {
		err = NewError("format to date clock ex check error: time=%v, err=%v", s, err)
	}
	return
}

//go:linkname now time.now
func now() (sec int64, nsec int32)

//返回秒级时间戳
func Unix() int64 {
	sec, _ := now()
	return sec
}

//返回毫秒级时间戳
func UnixMs() int64 {
	sec, nsec := now()
	return sec*1000 + int64(nsec/1000000)
}

//返回纳秒级时间戳
func UnixNano() int64 {
	sec, nsec := now()
	return sec*1e9 + int64(nsec)
}

//@description: 返回时间戳所在的时间是周几, 星期1为一周的开始
//@param:       unix int64 "秒级时间戳"
//@param:       zone TimeZone "时区"
//@return:      week int "星期(1-7)"
func UnixWeekdayA(unix int64, zone TimeZone) (week int) {
	unixLocal := unix + zone.Offset()
	if unixLocal < 0 {
		nSecond := int(unixLocal%weekSec+weekSec) % weekSec
		week = nSecond/daySec + 4
	} else {
		week = int(unixLocal%weekSec/daySec + 4)
	}
	if week > 7 {
		week = week - 7
	}
	return
}

//@description: 返回时间戳所在的时间是周几, 星期天为一周的开始
//@param:       unix int64 "秒级时间戳"
//@param:       zone TimeZone "时区"
//@return:      week int "星期(0-6)"
func UnixWeekdayB(unix int64, zone TimeZone) (week int) {
	week = UnixWeekdayA(unix, zone)
	if week == 7 {
		week = 0
	}
	return
}

//@description: 返回时间戳年中的星期数, 星期1为一周的开始
//@param:       unix int64 "秒级时间戳"
//@param:       zone TimeZone "时区"
//@return:      int(0-53) "第几周"
func UnixYearWeekNumA(unix int64, zone TimeZone) int {
	start := UnixYearZeroHour(unix, zone)
	nSecond := int(unix - start)
	nSecond += UnixWeekdayA(start, zone) * daySec
	return nSecond / weekSec
}

//@description: 返回时间戳年中的星期数, 星期天为一周的开始
//@param:       unix int64 "秒级时间戳"
//@param:       zone TimeZone "时区"
//@return:      int(0-53) "第几周"
func UnixYearWeekNumB(unix int64, zone TimeZone) int {
	start := UnixYearZeroHour(unix, zone)
	nSecond := int(unix - start)
	nSecond += UnixWeekdayB(start, zone) * daySec
	return nSecond / weekSec
}

//@description: 返回时间戳1970年1月1日以来的天数
//@param:       unix int64 "秒级时间戳"
//@param:       zone TimeZone "时区"
//@return:      int64 "天数"
func UnixDayNumber(unix int64, zone TimeZone) int64 {
	return (unix + zone.Offset()) / int64(daySec)
}

//@description: 返回时间戳所在年的1月1日0时的秒级时间戳
//@param:       unix int64 "秒级时间戳"
//@param:       zone TimeZone "时区"
//@return:      int64 "秒级时间戳"
func UnixYearZeroHour(unix int64, zone TimeZone) int64 {
	_, _, _, _, _, _, yDay, daySecond := UnixToDateClock(unix, zone)
	unixLocal := unix + zone.Offset()
	return unixLocal - int64((yDay-1)*daySec+daySecond) - zone.Offset()
}

//@description: 返回时间戳所在月1日0时的秒级时间戳
//@param:       unix int64 "秒级时间戳"
//@param:       zone TimeZone "时区"
//@return:      int64 "秒级时间戳"
func UnixMonthZeroHour(unix int64, zone TimeZone) int64 {
	year, month, _, _, _, _, _, _ := UnixToDateClock(unix, zone)
	unixMon, _, _, _ := DateClockToUnix(year, month, 1, 0, 0, 0, zone)
	return unixMon
}

//@description: 返回时间戳当天0时的秒级时间戳
//@param:       unix int64 "秒级时间戳"
//@param:       zone TimeZone "时区"
//@return:      int64 "秒级时间戳"
func UnixDayZeroHour(unix int64, zone TimeZone) int64 {
	unixLocal := unix + zone.Offset()
	if unixLocal < 0 {
		nSecond := (unixLocal%daySec + daySec) % daySec
		return unixLocal - nSecond - zone.Offset()
	}
	return unixLocal - unixLocal%daySec - zone.Offset()
}

//@description: 返回时间戳本小时的0分的秒级时间戳
//@param:       unix int64 "秒级时间戳"
//@return:      int64 "秒级时间戳"
func UnixHourZeroMin(unix int64) int64 {
	if unix < 0 {
		nSecond := (unix%hourSec + hourSec) % daySec
		return unix - nSecond
	}
	return unix - unix%hourSec

	//unixLocal := unix + zone.offset
	//if unixLocal < 0{
	//	nSecond := (unixLocal % hourSec + hourSec) % daySec
	//	return unixLocal - nSecond - zone.offset
	//}
	//return unixLocal - unixLocal % hourSec - zone.offset
}

//@description: 返回时间戳当天特定时间的秒级时间戳
//@param:       unix int64 "秒级时间戳"
//@param:       hour, min, sec int "时,分,秒"
//@param:       zone TimeZone "时区"
//@return:      int64 "秒级时间戳"
func UnixThisDay(unix int64, hour, min, sec int, zone TimeZone) (int64, error) {
	if err := checkClock(hour, min, sec); err != nil {
		return 0, err
	}
	start := UnixDayZeroHour(unix, zone)
	return start + int64(hour*hourSec+min*minSec+sec), nil
}

//@description: 返回时间戳后N天特定时间的秒级时间戳
//@param:       unix int64 "秒级时间戳"
//@param:       days, hour, min, sec int "天数,时,分,秒"
//@param:       zone TimeZone "时区"
//@return:      int64 "秒级时间戳"
func UnixDayZeroHourNext(unix int64, days, hour, min, sec int, zone TimeZone) (int64, error) {
	if err := checkClock(hour, min, sec); err != nil {
		return 0, err
	}
	start := UnixDayZeroHour(unix, zone)
	return start + int64(days*daySec+hour*hourSec+min*minSec+sec), nil
}

//@description: 返回时间戳下一周的星期几的秒级时间戳(星期1为周的开始)
//@param:       unix int64 "秒级时间戳"
//@param:       week, hour, min, sec int "星期几(1-7),时,分,秒"
//@param:       zone TimeZone "时区"
//@return:      int64 "秒级时间戳"
//@return:      error "错误信息"
func UnixNextWeekDayA(unix int64, week int, hour, min, sec int, zone TimeZone) (int64, error) {
	if week < 1 || week > 7 {
		return 0, NewError("week out of range(1,7): %v", week)
	}
	w := UnixWeekdayA(unix, zone)
	days := week - w
	return UnixDayZeroHour(unix, zone) + weekSec + int64(days)*daySec, nil
}

//@description: 返回时间戳下一周的星期几的秒级时间戳(星期天为周的开始)
//@param:       unix int64 "秒级时间戳"
//@param:       week, hour, min, sec int "星期几(0-6),时,分,秒"
//@param:       zone TimeZone "时区"
//@return:      int64 "秒级时间戳"
//@return:      error "错误信息"
func UnixNextWeekDayB(unix int64, week, hour, min, sec int, zone TimeZone) (int64, error) {
	if week < 0 || week > 6 {
		return 0, NewError("week out of range(0, 6): %v", week)
	}
	w := UnixWeekdayB(unix, zone)
	days := week - w
	return UnixDayZeroHour(unix, zone) + weekSec + int64(days)*daySec, nil
}

//@description: 返回时间戳下一个最近的星期几的秒级时间戳(星期1为周的开始)
//@param:       unix int64 "秒级时间戳"
//@param:       week, hour, min, sec int "星期几(1-7),时,分,秒"
//@param:       zone TimeZone "时区"
//@return:      int64 "秒级时间戳"
//@return:      error "错误信息"
func UnixFutureWeekDayA(unix int64, week, hour, min, sec int, zone TimeZone) (int64, error) {
	if week < 1 || week > 7 {
		return 0, NewError("week out of range(1,7): %v", week)
	}
	if err := checkClock(hour, min, sec); err != nil {
		return 0, err
	}

	w := UnixWeekdayA(unix, zone)
	days := week - w
	if days > 0 {
		return UnixDayZeroHour(unix, zone) + int64(days)*daySec + int64(hour*hourSec+min*minSec+sec), nil
	} else {
		return UnixDayZeroHour(unix, zone) + weekSec + int64(days)*daySec + int64(hour*hourSec+min*minSec+sec), nil
	}
}

//@description: 返回时间戳下一个最近的星期几的秒级时间戳(星期天为周的开始)
//@param:       unix int64 "秒级时间戳"
//@param:       week, hour, min, sec int "星期几(0-6),时,分,秒"
//@param:       zone TimeZone "时区"
//@return:      int64 "秒级时间戳"
//@return:      error "错误信息"
func UnixFutureWeekDayB(unix int64, week, hour, min, sec int, zone TimeZone) (int64, error) {
	if week < 0 || week > 6 {
		return 0, NewError("week out of range(0, 6): %v", week)
	}
	if err := checkClock(hour, min, sec); err != nil {
		return 0, err
	}

	w := UnixWeekdayB(unix, zone)
	days := week - w
	if days > 0 {
		return UnixDayZeroHour(unix, zone) + int64(days)*daySec + int64(hour*hourSec+min*minSec+sec), nil
	} else {
		return UnixDayZeroHour(unix, zone) + weekSec + int64(days)*daySec + int64(hour*hourSec+min*minSec+sec), nil
	}
}

//返回时间戳的 格式化日期时间字符串
func UnixToFormat(unix int64, zone TimeZone, formatter string) string {
	year, month, day, hour, min, sec, _, _ := UnixToDateClock(unix, zone)
	return DateClockToFormat(year, month, day, hour, min, sec, formatter)
}

//返回时间戳的 标准日期时间字符串
func UnixToYmdHMS(unix int64, zone TimeZone) string {
	return UnixToFormat(unix, zone, formatterYmdHMS)
}
