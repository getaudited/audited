package domain

import "time"

type Export struct {
	ID        string
	State     string
	Url       string
	Targets   []string
	Actors    []string
	Actions   []string
	RangeFrom time.Time
	RangeTo   time.Time
	CreatedAt time.Time
}
