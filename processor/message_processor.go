package processor

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/vithsutra/biometric-project-message-processor/models"
)

type messageProcessor struct {
	messageQueue     chan mqtt.Message
	mqttClient       mqtt.Client
	dbRepo           models.DeviceDatabseInterface
	workerNodesCount uint32
}

func NewMessageProcessor(
	mqttClient mqtt.Client,
	dbRepo models.DeviceDatabseInterface,
	workerNodesCount uint32,
	queueBufferSize uint32,
) *messageProcessor {
	return &messageProcessor{
		messageQueue:     make(chan mqtt.Message, queueBufferSize),
		mqttClient:       mqttClient,
		dbRepo:           dbRepo,
		workerNodesCount: workerNodesCount,
	}
}

func (p *messageProcessor) publishResponse(client mqtt.Client, deviceId string, payload []byte) {
	if client == nil || !client.IsConnected() {
		log.Println("cannot publish response: MQTT client not connected, DeviceId:", deviceId)
		return
	}
	token := client.Publish(deviceId, 1, false, payload)
	go func() {
		if token.WaitTimeout(3*time.Second) && token.Error() != nil {
			log.Println("failed to publish to device:", deviceId, "Error:", token.Error().Error())
		}
	}()
}

func (p *messageProcessor) processMessage(c mqtt.Client, message mqtt.Message) {
	if message == nil {
		return
	}
	topicArr := strings.Split(message.Topic(), "/")

	if len(topicArr) > 3 {
		deviceId := strings.TrimSpace(topicArr[0])
		if deviceId == "" {
			return
		}
		messageType := strings.TrimSpace(topicArr[2])
		switch messageType {
		case "connection":
			p.processDeviceConnectionRequest(c, deviceId, message.Payload())
		case "disconnection":
			p.processDeviceDisconnectionRequest(c, deviceId, message.Payload())
		case "deletesync":
			p.processDeviceDeleteSyncRequest(c, deviceId, message.Payload())
		case "deletesyncack":
			p.processDeviceDeleteSyncAckRequest(c, deviceId, message.Payload())
		case "insertsync":
			p.processDeviceInsertSyncRequest(c, deviceId, message.Payload())
		case "insertsyncack":
			p.processDeviceInsertSyncAckRequest(c, deviceId, message.Payload())
		case "attendance":
			p.processAttendanceRequest(c, deviceId, message.Payload())
		}
	}
}

func (p *messageProcessor) Start() {
	for i := 0; i < int(p.workerNodesCount); i++ {
		go func(workerID int) {
			for m := range p.messageQueue {
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("[PANIC RECOVERED] Worker %d recovered from panic: %v\n", workerID, r)
						}
					}()
					p.processMessage(p.mqttClient, m)
				}()
			}
		}(i)
	}
}

func (p *messageProcessor) Push(message mqtt.Message) {
	if message == nil {
		return
	}
	select {
	case p.messageQueue <- message:
	case <-time.After(500 * time.Millisecond):
		log.Println("[WARNING] messageQueue full, dropped message to avoid MQTT deadlock! Topic:", message.Topic())
	}
}

func (p *messageProcessor) processDeviceConnectionRequest(client mqtt.Client, deviceId string, message []byte) {
	deviceExists, err := p.dbRepo.CheckDeviceExists(deviceId)

	if err != nil {
		log.Println("error occurred with database while checking device exists, Device Id: ", deviceId, " Error: ", err.Error())
		response := models.ConnectionUpdateResponse{
			MessageType: 1,
			ErrorStatus: 1,
		}

		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	if !deviceExists {
		log.Println("connection request from the invalid device, Device Id: ", deviceId)

		response := models.ConnectionUpdateResponse{
			MessageType: 1,
			ErrorStatus: 1,
		}

		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	if err := p.dbRepo.UpdateDeviceStatus(deviceId, true); err != nil {
		log.Println("error occurred with database while updating the connection status, Device Id: ", deviceId, " Error: ", err.Error())
		response := models.ConnectionUpdateResponse{
			MessageType: 1,
			ErrorStatus: 1,
		}
		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}
	response := models.ConnectionUpdateResponse{
		MessageType: 1,
		ErrorStatus: 0,
	}
	responseJson, _ := json.Marshal(response)
	p.publishResponse(client, deviceId, responseJson)
}

func (p *messageProcessor) processDeviceDisconnectionRequest(client mqtt.Client, deviceId string, message []byte) {
	if err := p.dbRepo.UpdateDeviceStatus(deviceId, false); err != nil {
		log.Println("error occurred with database while updating the disconnection status, Device Id: ", deviceId, " Error: ", err.Error())
	}
}

func (p *messageProcessor) processDeviceDeleteSyncRequest(client mqtt.Client, deviceId string, message []byte) {

	exists, err := p.dbRepo.CheckStudentsExistsInDeletes(deviceId)

	if err != nil {
		log.Println("error occurred with database while checking students exists in deletes, Device Id: ", deviceId, " Error: ", err.Error())

		response := models.DeleteSyncResponse{
			MessageType:   2,
			ErrorStatus:   1,
			StudentsEmpty: 0,
			StudentId:     0,
		}

		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	if !exists {
		response := models.DeleteSyncResponse{
			MessageType:   2,
			ErrorStatus:   0,
			StudentsEmpty: 1,
			StudentId:     0,
		}

		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	studentId, err := p.dbRepo.GetStudentFromDeletes(deviceId)

	if err != nil {
		log.Println("error occurred with database while getting student from deletes, Device Id: ", deviceId, " Error: ", err.Error())

		response := models.DeleteSyncResponse{
			MessageType:   2,
			ErrorStatus:   1,
			StudentsEmpty: 0,
			StudentId:     0,
		}

		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	studentIdInt, _ := strconv.Atoi(studentId)

	response := models.DeleteSyncResponse{
		MessageType:   2,
		ErrorStatus:   0,
		StudentsEmpty: 0,
		StudentId:     uint16(studentIdInt),
	}

	responseJson, _ := json.Marshal(response)
	p.publishResponse(client, deviceId, responseJson)

}

func (p *messageProcessor) processDeviceDeleteSyncAckRequest(client mqtt.Client, deviceId string, message []byte) {

	req := new(models.DeleteSyncAckRequest)

	if err := json.Unmarshal(message, req); err != nil {
		log.Println("invalid json format in the delete sync ack request, Device Id: ", deviceId, " Error: ", err.Error())

		response := models.DeleteSyncAckResponse{
			MessageType: 3,
			ErrorStatus: 1,
		}

		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	if err := p.dbRepo.DeleteStudentFromDeletes(deviceId, strconv.Itoa(int(req.StudentId))); err != nil {
		log.Println("error occurred with database while deleting the student from deletes, Device Id: ", deviceId, " Error: ", err.Error())
		response := models.DeleteSyncAckResponse{
			MessageType: 3,
			ErrorStatus: 1,
		}
		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	response := models.DeleteSyncAckResponse{
		MessageType: 3,
		ErrorStatus: 0,
	}
	responseJson, _ := json.Marshal(response)
	p.publishResponse(client, deviceId, responseJson)

}

func (p *messageProcessor) processDeviceInsertSyncRequest(client mqtt.Client, deviceId string, message []byte) {

	exists, err := p.dbRepo.CheckStudentsExistsInInserts(deviceId)

	if err != nil {
		log.Println("error occurred with database while checking student exists in inserts, Device Id: ", deviceId, " Error: ", err.Error())
		response := models.InsertSyncResponse{
			MessageType: 4,
			ErrorStatus: 1,
		}
		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	if !exists {
		response := models.InsertSyncResponse{
			MessageType:   4,
			ErrorStatus:   0,
			StudentsEmpty: 1,
		}
		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	studentId, fingerprintData, err := p.dbRepo.GetStudentFromInserts(deviceId)

	if err != nil {
		log.Println("error occurred with database while getting student from inserts, Device Id: ", deviceId, " Error: ", err.Error())
		response := models.InsertSyncResponse{
			MessageType: 4,
			ErrorStatus: 1,
		}
		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	studentIdInt, _ := strconv.Atoi(studentId)

	response := models.InsertSyncResponse{
		MessageType:     4,
		ErrorStatus:     0,
		StudentsEmpty:   0,
		StudentId:       uint16(studentIdInt),
		FingerPrintData: fingerprintData,
	}

	responseJson, _ := json.Marshal(response)
	p.publishResponse(client, deviceId, responseJson)
}

func (p *messageProcessor) processDeviceInsertSyncAckRequest(client mqtt.Client, deviceId string, message []byte) {
	req := new(models.InsertSyncAckRequest)

	if err := json.Unmarshal(message, req); err != nil {
		log.Println("error occurred while decoding json insert sync ack message, Device Id: ", deviceId, " Error: ", err.Error())
		response := models.InsertSyncAckResponse{
			MessageType: 5,
			ErrorStatus: 1,
		}
		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	if err := p.dbRepo.DeleteStudentFromInserts(deviceId, strconv.Itoa(int(req.StudentId))); err != nil {
		log.Println("error occurred while deleting the student from inserts, Device Id: ", deviceId, " Error: ", err.Error())

		response := models.InsertSyncAckResponse{
			MessageType: 5,
			ErrorStatus: 1,
		}

		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	response := models.InsertSyncAckResponse{
		MessageType: 5,
		ErrorStatus: 0,
	}

	responseJson, _ := json.Marshal(response)
	p.publishResponse(client, deviceId, responseJson)
}

func (p *messageProcessor) processAttendanceRequest(client mqtt.Client, deviceId string, message []byte) {

	req := new(models.UpdateAttendanceRequest)

	if err := json.Unmarshal(message, req); err != nil {
		log.Println("error occurred while decoding the json in update attendance request, DeviceId:", deviceId, " Error: ", err.Error())

		response := models.UpdateAttendanceResponse{
			MessageType: 6,
			ErrorStatus: 1,
			Index:       req.Index,
		}

		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	studentId, err := p.dbRepo.GetStudentId(deviceId, strconv.Itoa(int(req.StudentUnitId)))

	if err != nil {
		log.Println("error occurred while updating the student attendance, DeviceId: ", deviceId, "StudentUnitId: ", req.StudentUnitId, " Error: ", err.Error())
		response := models.UpdateAttendanceResponse{
			MessageType: 6,
			ErrorStatus: 1,
			Index:       req.Index,
		}

		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	var t time.Time
	t, err = time.Parse("2006-01-02T15:04:05", req.TimeStamp)
	if err != nil {
		t, err = time.Parse("2006-01-02 15:04:05", req.TimeStamp)
	}

	if err != nil {
		log.Println("error occurred while parsing the attendance timestamp, DeviceId: ", deviceId, "StudentUnitId: ", req.StudentUnitId, " Error: ", err.Error())
		response := models.UpdateAttendanceResponse{
			MessageType: 6,
			ErrorStatus: 1,
			Index:       req.Index,
		}

		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	date := t.Format("2006-01-02")

	tm := t.Format("15:04")

	isLogout, err := p.dbRepo.CheckLoginOrLogout(studentId, date)

	if err != nil {
		log.Println("error occurred with database while checking attedance login or logout, DeviceId: ", deviceId, "StudentUnitId: ", req.StudentUnitId, " Error: ", err.Error())
		response := models.UpdateAttendanceResponse{
			MessageType: 6,
			ErrorStatus: 1,
			Index:       req.Index,
		}

		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
		return
	}

	if isLogout {
		if err := p.dbRepo.UpdateAttendanceLog(studentId, date, tm); err != nil {
			log.Println("error occurred while updating the student attendance, DeviceId: ", deviceId, "StudentUnitId: ", req.StudentUnitId, " Error: ", err.Error())
			response := models.UpdateAttendanceResponse{
				MessageType: 6,
				ErrorStatus: 1,
				Index:       req.Index,
			}

			responseJson, _ := json.Marshal(response)
			p.publishResponse(client, deviceId, responseJson)
			return
		}

		response := models.UpdateAttendanceResponse{
			MessageType: 6,
			ErrorStatus: 0,
			Index:       req.Index,
		}
		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
	} else {
		att := new(models.Attendance)

		att.StudentId = studentId
		att.Date = date
		att.Login = tm
		att.Logout = "25:00"

		if err := p.dbRepo.InsertAttendanceLog(att); err != nil {
			log.Println("error occurred with database while inserting the attendance, DeviceId: ", deviceId, "StudentUnitId: ", req.StudentUnitId, " Error: ", err.Error())
			response := models.UpdateAttendanceResponse{
				MessageType: 6,
				ErrorStatus: 1,
				Index:       req.Index,
			}

			responseJson, _ := json.Marshal(response)
			p.publishResponse(client, deviceId, responseJson)
			return
		}

		response := models.UpdateAttendanceResponse{
			MessageType: 6,
			ErrorStatus: 0,
			Index:       req.Index,
		}
		responseJson, _ := json.Marshal(response)
		p.publishResponse(client, deviceId, responseJson)
	}

}
