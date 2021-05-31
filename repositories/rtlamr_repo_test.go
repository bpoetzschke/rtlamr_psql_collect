package repositories

import (
	"database/sql"
	"testing"
	"time"

	"github.com/bpoetzschke/rtlamr_psql_collect/database"
	"github.com/bpoetzschke/rtlamr_psql_collect/models"
	"github.com/stretchr/testify/suite"
)

type RTLAMRRepoSuite struct {
	suite.Suite

	db   *sql.DB
	repo RTLAMRRepo
}

func (s *RTLAMRRepoSuite) SetupSuite() {
	db, err := database.Init()
	s.Require().NoError(err)

	s.db = db
	s.repo = NewRTLAMRRepo(s.db)
}

func (s *RTLAMRRepoSuite) TearDownTest() {
	_, err := s.db.Exec("TRUNCATE TABLE rtlamr_data;")
	s.Require().NoError(err)
}

func (s *RTLAMRRepoSuite) TearDownSuite() {
	err := s.db.Close()
	s.Require().NoError(err)
}

func (s RTLAMRRepoSuite) TestGetLastReading() {
	_, found, err := s.repo.GetLastReading("1234")
	s.Require().NoError(err)
	s.Require().False(found)

	_, err = s.db.Exec(`INSERT INTO rtlamr_data (meter_id, meter_type, current_reading, difference, created_at)
					 VALUES ('1234', '7', '100', '5', '2021-05-31 12:00:00');`)
	s.Require().NoError(err)

	expectedTime := time.Date(2021, time.May, 31, 12, 0, 0, 0, time.UTC)

	data, found, err := s.repo.GetLastReading("1234")
	s.Require().NoError(err)
	s.Require().True(found)
	s.Require().EqualValues("1234", data.MeterID)
	s.Require().EqualValues("7", data.MeterType)
	s.Require().EqualValues(100, data.CurrentReading)
	s.Require().EqualValues(5, data.Difference)
	s.Require().EqualValues(expectedTime.Nanosecond(), data.CreatedAt.Nanosecond())
}

func (s RTLAMRRepoSuite) TestStoreReading() {
	data := models.RTLAMRData{
		CreatedAt:      time.Now().UTC(),
		MeterID:        "1234",
		MeterType:      "7",
		CurrentReading: 100,
		Difference:     5,
	}

	stored, err := s.repo.StoreReading(data)
	s.Require().NoError(err)
	s.Require().True(stored)

	stored, err = s.repo.StoreReading(data)
	s.Require().NoError(err)
	s.Require().False(stored)

	storedData, found, err := s.repo.GetLastReading("1234")
	s.Require().NoError(err)
	s.Require().True(found)
	s.Require().EqualValues(data.MeterID, storedData.MeterID)
	s.Require().EqualValues(data.MeterType, storedData.MeterType)
	s.Require().EqualValues(data.CurrentReading, storedData.CurrentReading)
	s.Require().EqualValues(data.Difference, storedData.Difference)
	s.Require().EqualValues(data.CreatedAt.Nanosecond(), storedData.CreatedAt.Nanosecond())
}

func TestRTLAMRRepo(t *testing.T) {
	suite.Run(t, &RTLAMRRepoSuite{})
}
