package job

type Option func(*Job)

func WithConcurrencyLimit(limit int) Option {
	return func(j *Job) {
		j.LimitConcurrency = true
		j.MaxConcurrency = limit
	}
}

func WithRunLimit(limit int) Option {
	return func(j *Job) {
		j.LimitRuns = true
		j.MaxRuns = limit
	}
}

func WithDisabled() Option {
	return func(j *Job) {
		j.Enabled = false
	}
}
