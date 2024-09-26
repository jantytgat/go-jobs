package cron

type position int

func (p position) String() string {
	return [...]string{"second", "minute", "hour", "day", "month", "weekday", "year"}[p]
}

const (
	positionSecond position = iota
	positionMinute
	positionHour
	positionDay
	positionMonth
	positionWeekday
	positionYear
)
