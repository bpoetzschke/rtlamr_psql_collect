package rtlamrclient

import (
	"fmt"
	"time"

	"github.com/bpoetzschke/rtlamr_psql_collect/models"
)

type ClientData struct {
	Time    string            `json:"time"`
	Type    string            `json:"type"`
	Message ClientDataMessage `json:"message"`
}

type ClientDataMessage struct {
	ID          int64 `json:"id"`
	Type        int64 `json:"type"`
	Consumption int64 `json:"consumption"`
}

func (cd ClientData) ToRTLAMRData() (models.RTLAMRData, error) {
	parsedTime, err := time.Parse(time.RFC3339, cd.Time)
	if err != nil {
		return models.RTLAMRData{}, err
	}

	return models.RTLAMRData{
		CreatedAt:      parsedTime.UTC(),
		MeterID:        fmt.Sprintf("%d", cd.Message.ID),
		MeterType:      fmt.Sprintf("%d", cd.Message.Type),
		CurrentReading: cd.Message.Consumption,
		Difference:     0,
	}, nil
}
