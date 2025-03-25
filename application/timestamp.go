package application

import (
	"fmt"
	"time"
)

type Timestamp time.Time

func (t Timestamp) String() string {
	v := time.Time(t)
	return fmt.Sprintf("%02d:%02d:%02d,%.03d", v.Hour(), v.Minute(), v.Second(), v.Nanosecond()/1000000)
}

func (t *Timestamp) Add(d time.Duration) Timestamp {
	v := time.Time(*t)
	return Timestamp(v.Add(d))
}
