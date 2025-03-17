package models

type ConnectionUpdateResponse struct {
	MessageType string `json:"mty"`
	ErrorStatus string `json:"est"`
}
type DeleteSyncResponse struct {
	MessageType   string `json:"mty"`
	ErrorStatus   string `json:"est"`
	StudentsEmpty string `json:"ste"`
	StudentId     string `json:"sid"`
}

type DeleteSyncAckRequest struct {
	StudentId string `json:"sid"`
}

type DeleteSyncAckResponse struct {
	MessageType string `json:"mty"`
	ErrorStatus string `json:"est"`
}

type InsertSyncResponse struct {
	MessageType     string `json:"mty"`
	ErrorStatus     string `json:"est"`
	StudentsEmpty   string `json:"ste"`
	StudentId       string `json:"sid"`
	FingerPrintData string `json:"fpd"`
}

type InsertSyncAckRequest struct {
	StudentId string `json:"sid"`
}

type InsertSyncAckResponse struct {
	MessageType string `json:"mty"`
	ErrorStatus string `json:"est"`
}

type UpdateAttendanceRequest struct {
	StudentId string `json:"sid"`
	TimeStamp string `json:"tmstmp"`
}

type UpdateAttendanceResponse struct {
	MessageType string `json:"mty"`
	ErrorStatus string `json:"est"`
}

type Attendance struct {
	DeviceId  string `json:"did"`
	StudentId string `json:"sid"`
	TimeStamp string `json:"tmstmp"`
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
}

type DeviceQueueInterface interface {
	PublishAttendance(*Attendance) error
}
