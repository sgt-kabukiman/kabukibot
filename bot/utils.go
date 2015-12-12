package bot

import (
	"fmt"
	"regexp"
)

import "strconv"
import "strings"
import "time"

const (
	ONE_SECOND = 1
	ONE_MINUTE = 60 * ONE_SECOND
	ONE_HOUR   = 60 * ONE_MINUTE
	ONE_DAY    = 24 * ONE_HOUR
	ONE_WEEK   = 7 * ONE_DAY
)

var oneDay = 24 * time.Hour
var handyDay = regexp.MustCompile(`[0-9]+d`)
var groupMatcher = regexp.MustCompile(`([0-9]+)([a-z]+)`)

var extUnits = map[string][]string{
	"d": []string{"day", "days"},
	"h": []string{"hour", "hours"},
	"m": []string{"minute", "minutes"},
	"s": []string{"second", "seconds"},
}

func ParseDuration(val string, min *time.Duration, max *time.Duration) *time.Duration {
	val = handyDay.ReplaceAllStringFunc(val, func(match string) string {
		num, err := strconv.Atoi(match[:len(match)-1])
		if err == nil {
			return strconv.Itoa(num*24) + "h"
		}

		return match
	})

	parsed, err := time.ParseDuration(val)
	if err != nil {
		return min
	}

	if min != nil && parsed < *min {
		return min
	}

	if max != nil && parsed > *max {
		return max
	}

	return &parsed
}

func FormatDuration(d time.Duration, extended bool) string {
	result := ""

	if d >= oneDay {
		days := int(d.Hours()) / int(oneDay.Hours())
		d -= time.Duration(float64(days) * 24 * float64(time.Hour))

		result = strconv.Itoa(days) + "d"
	}

	duration := d.String()
	if duration != "" && duration != "0" {
		result += duration
	}

	// remove trailing zero values, like in "1m0s"
	if !extended {
		shortened := ""

		for _, match := range groupMatcher.FindAllStringSubmatch(result, -1) {
			if match[1] != "0" {
				shortened += match[0]
			}
		}

		return shortened
	} else {
		var shortened []string

		for _, match := range groupMatcher.FindAllStringSubmatch(result, -1) {
			if match[1] != "0" {
				unit, okay := extUnits[match[2]]
				if okay {
					var u string

					if match[1] == "1" {
						u = unit[0]
					} else {
						u = unit[1]
					}

					shortened = append(shortened, match[1]+" "+u)
				} else {
					shortened = append(shortened, match[0])
				}
			}
		}

		return HumanJoin(shortened, ", ")
	}
}

func twodigit(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}

	return strconv.Itoa(n)
}

func plural(n int, word string) string {
	res := fmt.Sprintf("%d %s", n, word)

	if n != 1 {
		res = res + "s"
	}

	return res
}

func FormatDateAsSQL(t time.Time) string {
	return t.Format("2006-01-02")
}

func SecondsToTime(seconds int) string {
	return secondsToTime(seconds, false)
}

func SecondsToTimeCompact(seconds int) string {
	return secondsToTime(seconds, true)
}

func secondsToTime(seconds int, compact bool) string {
	weeks := seconds / ONE_WEEK
	seconds -= (weeks * ONE_WEEK)

	days := seconds / ONE_DAY
	seconds -= (days * ONE_DAY)

	hours := seconds / ONE_HOUR
	seconds -= (hours * ONE_HOUR)

	minutes := seconds / ONE_MINUTE
	seconds -= (minutes * ONE_MINUTE)

	list := make([]string, 0)

	if compact {
		if weeks > 0 {
			list = append(list, twodigit(weeks)+"w")
		}
		if len(list) > 0 || days > 0 {
			list = append(list, twodigit(days)+"d")
		}
		if len(list) > 0 || hours > 0 {
			list = append(list, twodigit(hours)+"h")
		}
		if len(list) > 0 || minutes > 0 {
			list = append(list, twodigit(minutes)+"m")
		}
		if len(list) > 0 || seconds > 0 {
			list = append(list, twodigit(seconds)+"s")
		}

		return strings.Join(list, ":")
	}

	if weeks > 0 {
		list = append(list, plural(weeks, "week"))
	}
	if days > 0 {
		list = append(list, plural(days, "day"))
	}
	if hours > 0 {
		list = append(list, plural(hours, "hour"))
	}
	if minutes > 0 {
		list = append(list, plural(minutes, "minute"))
	}
	if seconds > 0 {
		list = append(list, plural(seconds, "second"))
	}

	return HumanJoin(list, ", ")
}

func HumanJoin(list []string, glue string) string {
	if glue == "" {
		glue = ", "
	}

	l := len(list)

	switch l {
	case 0:
		return ""
	case 1:
		return list[0]
	default:
		return strings.Join(list[:(l-1)], glue) + " and " + list[l-1]
	}
}
