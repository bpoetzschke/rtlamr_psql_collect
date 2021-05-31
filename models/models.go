package models

import "time"

type RTLAMRData struct {
	CreatedAt      time.Time
	MeterID        string
	MeterType      string
	CurrentReading int64
	Difference     int64
}
