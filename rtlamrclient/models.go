package rtlamrclient

import (
	"fmt"
	"time"

	"github.com/bpoetzschke/rtlamr_psql_collect/models"
)

type ClientData struct {
	Time    time.Time         `json:"time"`
	Type    string            `json:"type"`
	Message ClientDataMessage `json:"message"`
}

type ClientDataMessage struct {
	ID          int64 `json:"id"`
	Type        int64 `json:"type"`
	Consumption int64 `json:"consumption"`
}

func (cd ClientData) ToRTLAMRData() models.RTLAMRData {
	return models.RTLAMRData{
		CreatedAt:      cd.Time,
		MeterID:        fmt.Sprintf("%d", cd.Message.ID),
		MeterType:      fmt.Sprintf("%d", cd.Message.Type),
		CurrentReading: cd.Message.Consumption,
		Difference:     0,
	}
}
