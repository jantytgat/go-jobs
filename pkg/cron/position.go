package cron

const (
	positionSecond position = iota
	positionMinute
	positionHour
	positionDay
	positionMonth
	positionWeekday
	positionYear
)

var positionStrings = []string{"second", "minute", "hour", "day", "month", "weekday", "year"}

type position int

func (p position) String() string {
	return positionStrings[p]
}
