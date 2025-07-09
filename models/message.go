package models

type ConnectionUpdateResponse struct {
	MessageType uint8 `json:"mty"`
	ErrorStatus uint8 `json:"est"`
}
type DeleteSyncResponse struct {
	MessageType   uint8  `json:"mty"`
	ErrorStatus   uint8  `json:"est"`
	StudentsEmpty uint8  `json:"ste"`
	StudentId     uint16 `json:"sid"`
}

type DeleteSyncAckRequest struct {
	StudentId uint16 `json:"sid"`
}

type DeleteSyncAckResponse struct {
	MessageType uint8 `json:"mty"`
	ErrorStatus uint8 `json:"est"`
}

type InsertSyncResponse struct {
	MessageType     uint8  `json:"mty"`
	ErrorStatus     uint8  `json:"est"`
	StudentsEmpty   uint8  `json:"ste"`
	StudentId       uint16 `json:"sid"`
	FingerPrintData string `json:"fpd"`
}

type InsertSyncAckRequest struct {
	StudentId uint16 `json:"sid"`
}

type InsertSyncAckResponse struct {
	MessageType uint8 `json:"mty"`
	ErrorStatus uint8 `json:"est"`
}

type UpdateAttendanceRequest struct {
	// MessageId     string `json:"mid"`
	StudentUnitId uint16 `json:"sid"`
	Index         uint32 `json:"index"`
	TimeStamp     string `json:"tmstmp"`
}

type UpdateAttendanceResponse struct {
	MessageType uint8  `json:"mty"`
	ErrorStatus uint8  `json:"est"`
	Index       uint32 `json:"index"`
}

type Attendance struct {
	StudentId string
	Date      string
	Login     string
	Logout    string
}

type DeviceDatabseInterface interface {
	CheckDeviceExists(deviceId string) (bool, error)
	UpdateDeviceStatus(deviceId string, status bool) error
	CheckStudentsExistsInDeletes(deviceId string) (bool, error)
	GetStudentFromDeletes(deviceId string) (string, error)
	DeleteStudentFromDeletes(deviceId string, studentId string) error
	CheckStudentsExistsInInserts(deviceId string) (bool, error)
	GetStudentFromInserts(deviceId string) (string, string, error)
	DeleteStudentFromInserts(deviceId string, studentId string) error
	GetStudentId(unitId string, studentUnitId string) (string, error)
	CheckLoginOrLogout(studentId string, date string) (bool, error)
	InsertAttendanceLog(attendanceLog *Attendance) error
	UpdateAttendanceLog(studentId string, date string, logout string) error
}

type DeviceCacheInterface interface {
	CheckMessageDuplication(messageId string) (bool, error)
}
