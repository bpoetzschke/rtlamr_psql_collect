package repositories

import (
	"database/sql"

	"github.com/bpoetzschke/rtlamr_psql_collect/models"
)

type RTLAMRRepo interface {
	GetLastReading(meterID string) (data models.RTLAMRData, found bool, err error)
	StoreReading(data models.RTLAMRData) (stored bool, err error)
}

func NewRTLAMRRepo(db *sql.DB) RTLAMRRepo {
	return &rtlamrRepo{
		db: db,
	}
}

type rtlamrRepo struct {
	db *sql.DB
}

func (r *rtlamrRepo) GetLastReading(meterID string) (models.RTLAMRData, bool, error) {
	selectQuery := `SELECT meter_id, meter_type, current_reading, difference, created_at 
					FROM rtlamr_data WHERE meter_id=$1 ORDER BY created_at DESC LIMIT 1;`

	res := models.RTLAMRData{}

	row := r.db.QueryRow(selectQuery, meterID)
	err := row.Scan(
		&res.MeterID,
		&res.MeterType,
		&res.CurrentReading,
		&res.Difference,
		&res.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return models.RTLAMRData{}, false, nil
	}

	if err != nil {
		return models.RTLAMRData{}, false, err
	}

	return res, true, nil
}

func (r *rtlamrRepo) StoreReading(data models.RTLAMRData) (bool, error) {
	insertQuery := `INSERT INTO rtlamr_data (meter_id, meter_type, current_reading, difference, created_at)
					VALUES ($1, $2, $3, $4, $5)
					ON CONFLICT (meter_id, current_reading) DO NOTHING;`

	res, err := r.db.Exec(insertQuery, data.MeterID, data.MeterType, data.CurrentReading, data.Difference, data.CreatedAt)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected == 1, nil
}
