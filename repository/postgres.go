package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vithsutra/biometric-project-message-processor/models"
)

type postgresRepository struct {
	dbConn *pgxpool.Pool
}

func NewPostgresRepository(dbConn *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{
		dbConn,
	}
}

func (repo *postgresRepository) CheckDeviceExists(deviceId string) (bool, error) {
	query := `SELECT EXISTS ( SELECT 1 FROM biometric WHERE unit_id = $1 )`
	var exists bool
	err := repo.dbConn.QueryRow(context.Background(), query, deviceId).Scan(&exists)
	return exists, err
}

func (repo *postgresRepository) UpdateDeviceStatus(deviceId string, status bool) error {
	query := `UPDATE biometric SET online=$2 WHERE unit_id=$1`
	_, err := repo.dbConn.Exec(context.Background(), query, deviceId, status)
	return err
}

func (repo *postgresRepository) CheckStudentsExistsInDeletes(deviceId string) (bool, error) {
	query := `SELECT EXISTS ( SELECT 1 FROM deletes WHERE unit_id=$1 )`
	var exists bool
	err := repo.dbConn.QueryRow(context.Background(), query, deviceId).Scan(&exists)
	return exists, err
}

func (repo *postgresRepository) GetStudentFromDeletes(deviceId string) (string, error) {
	query := `SELECT student_unit_id FROM deletes WHERE unit_id=$1`
	var id string
	err := repo.dbConn.QueryRow(context.Background(), query, deviceId).Scan(&id)
	return id, err
}

func (repo *postgresRepository) DeleteStudentFromDeletes(deviceId string, studentId string) error {
	query := `DELETE FROM deletes WHERE unit_id=$1 AND student_unit_id=$2`
	_, err := repo.dbConn.Exec(context.Background(), query, deviceId, studentId)
	return err
}

func (repo *postgresRepository) CheckStudentsExistsInInserts(deviceId string) (bool, error) {
	query := `SELECT EXISTS ( SELECT 1 FROM inserts WHERE unit_id=$1 )`
	var exists bool
	err := repo.dbConn.QueryRow(context.Background(), query, deviceId).Scan(&exists)
	return exists, err
}

func (repo *postgresRepository) GetStudentFromInserts(deviceId string) (string, string, error) {
	query := `SELECT student_unit_id,fingerprint_data FROM inserts WHERE unit_id=$1`
	var id, fingerprint string
	err := repo.dbConn.QueryRow(context.Background(), query, deviceId).Scan(&id, &fingerprint)
	return id, fingerprint, err
}

func (repo *postgresRepository) DeleteStudentFromInserts(deviceId string, studentId string) error {
	query := `DELETE FROM inserts WHERE unit_id=$1 AND student_unit_id=$2`
	_, err := repo.dbConn.Exec(context.Background(), query, deviceId, studentId)
	return err
}

func (repo *postgresRepository) GetStudentId(unitId string, studentUnitId string) (string, error) {
	query := `SELECT student_id FROM fingerprintdata WHERE unit_id=$1 AND student_unit_id=$2`
	var studentId string
	err := repo.dbConn.QueryRow(context.Background(), query, unitId, studentUnitId).Scan(&studentId)
	return studentId, err
}

func (repo *postgresRepository) CheckLoginOrLogout(studentId string, date string) (bool, error) {
	query := `SELECT EXISTS ( SELECT 1 FROM attendance WHERE date=$1 AND student_id=$2 and logout=$3 )`
	var logStatus bool
	err := repo.dbConn.QueryRow(context.Background(), query, date, studentId, "25:00").Scan(&logStatus)
	return logStatus, err
}

func (repo *postgresRepository) InsertAttendanceLog(attendanceLog *models.Attendance) error {
	query := `INSERT INTO attendance (student_id,date,login,logout) VALUES ($1,$2,$3,$4)`

	_, err := repo.dbConn.Exec(
		context.Background(),
		query,
		attendanceLog.StudentId,
		attendanceLog.Date,
		attendanceLog.Login,
		attendanceLog.Logout,
	)

	return err
}

func (repo *postgresRepository) UpdateAttendanceLog(studentId string, date string, logout string) error {
	query := `UPDATE attendance SET logout=$4 WHERE student_id=$1 AND date=$2 AND logout=$3`
	_, err := repo.dbConn.Exec(
		context.Background(),
		query,
		studentId,
		date,
		"25:00",
		logout,
	)
	return err
}
