package bot

import "fmt"
import "math"
import "strconv"
import "strings"
import "time"

const (
	ONE_SECOND =  1
	ONE_MINUTE = 60 * ONE_SECOND
	ONE_HOUR   = 60 * ONE_MINUTE
	ONE_DAY    = 24 * ONE_HOUR
	ONE_WEEK   =  7 * ONE_DAY
)

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
	weeks   := seconds / ONE_WEEK
	seconds -= (weeks * ONE_WEEK)

	days    := seconds / ONE_DAY
	seconds -= (days * ONE_DAY)

	hours   := seconds / ONE_HOUR
	seconds -= (hours * ONE_HOUR)

	minutes := seconds / ONE_MINUTE
	seconds -= (minutes * ONE_MINUTE)

	list := make([]string, 0)

	if compact {
		if                  weeks > 0   { list = append(list, twodigit(weeks)   + "w") }
		if len(list) > 0 || days > 0    { list = append(list, twodigit(days)    + "d") }
		if len(list) > 0 || hours > 0   { list = append(list, twodigit(hours)   + "h") }
		if len(list) > 0 || minutes > 0 { list = append(list, twodigit(minutes) + "m") }
		if len(list) > 0 || seconds > 0 { list = append(list, twodigit(seconds) + "s") }

		return strings.Join(list, ":")
	}

	if weeks > 0   { list = append(list, plural(weeks,   "week"))   }
	if days > 0    { list = append(list, plural(days,    "day"))    }
	if hours > 0   { list = append(list, plural(hours,   "hour"))   }
	if minutes > 0 { list = append(list, plural(minutes, "minute")) }
	if seconds > 0 { list = append(list, plural(seconds, "second")) }

	return HumanJoin(list, ", ")
}

func SecondsToRunTime(seconds float32) string {
	hours   := int(seconds / ONE_HOUR)
	seconds -= float32(hours * ONE_HOUR)

	minutes := int(seconds / ONE_MINUTE)
	seconds -= float32(minutes * ONE_MINUTE)

	list := make([]string, 0)

	if                  hours > 0   { list = append(list, twodigit(hours))        }
	if len(list) > 0 || minutes > 0 { list = append(list, twodigit(minutes))      }
	if len(list) > 0 || seconds > 0 { list = append(list, twodigit(int(seconds))) }

	runtime := strings.TrimPrefix(strings.Join(list, ":"), "0")

	seconds -= float32(int(seconds) * ONE_SECOND)

	if seconds > 0.0001 {
		seconds *= 1000 // 0.12345 => 123.45
		seconds  = float32(math.Floor(float64(seconds + 0.5)))
		runtime += fmt.Sprintf(".%d", int(seconds))
	}

	return runtime
}

func HumanJoin(list []string, glue string) string {
	if glue == "" {
		glue = ", "
	}

	l := len(list)

	switch l {
		case 0:  return ""
		case 1:  return list[0]
		default: return strings.Join(list[:(l-1)], glue) + " and " + list[l-1]
	}
}
