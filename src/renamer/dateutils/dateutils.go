package dateutils

import (
	"fmt"
	"strconv"
	"time"
)

func GetMonth(date time.Time) string {
	switch date.Month() {
	case time.January:
		return "01"
	case time.February:
		return "02"
	case time.March:
		return "03"
	case time.April:
		return "04"
	case time.May:
		return "05"
	case time.June:
		return "06"
	case time.July:
		return "07"
	case time.August:
		return "08"
	case time.September:
		return "09"
	case time.October:
		return "10"
	case time.November:
		return "11"
	case time.December:
		return "12"
	}
	return "00"
}

func AddZero(num int) string {
	if num < 10 {
		return "0" + strconv.Itoa(num)
	} else {
		return strconv.Itoa(num)
	}

}

func DateToString(date time.Time) string {
	return 	fmt.Sprintf("%s:%s:%s %s:%s:%s", strconv.Itoa(date.Year()), GetMonth(date), AddZero(date.Day()), AddZero(date.Hour()), AddZero(date.Minute()), AddZero(date.Second()))
}
