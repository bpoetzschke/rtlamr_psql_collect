package repositories

import (
	"github.com/bpoetzschke/rtlamr_psql_collect/models"
	"github.com/stretchr/testify/mock"
)

type RTLAMRRepoMock struct {
	mock.Mock
}

func (m *RTLAMRRepoMock) GetLastReading(meterID string) (data models.RTLAMRData, found bool, err error) {
	called := m.Called(meterID)
	return called.Get(0).(models.RTLAMRData), called.Bool(1), called.Error(2)
}

func (m *RTLAMRRepoMock) StoreReading(data models.RTLAMRData) (stored bool, err error) {
	called := m.Called(data)
	return called.Bool(0), called.Error(1)
}
