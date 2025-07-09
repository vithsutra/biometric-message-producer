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
	cacheRepo        models.DeviceCacheInterface
	workerNodesCount uint32
}

func NewMessageProcessor(
	mqttClient mqtt.Client,
	dbRepo models.DeviceDatabseInterface,
	cacheRepo models.DeviceCacheInterface,
	workerNodesCount uint32,
	queueBufferSize uint32,
) *messageProcessor {
	return &messageProcessor{
		messageQueue:     make(chan mqtt.Message, queueBufferSize),
		mqttClient:       mqttClient,
		dbRepo:           dbRepo,
		cacheRepo:        cacheRepo,
		workerNodesCount: workerNodesCount,
	}
}

func (p *messageProcessor) processMessage(c mqtt.Client, message mqtt.Message) {
	topicArr := strings.Split(message.Topic(), "/")

	if len(topicArr) > 3 {
		deviceId := topicArr[0]
		deviceId = strings.Trim(deviceId, " ")
		messageType := topicArr[2]
		messageType = strings.Trim(messageType, " ")
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
		go func() {
			for m := range p.messageQueue {
				p.processMessage(p.mqttClient, m)
			}
		}()
	}
}

func (p *messageProcessor) Push(message mqtt.Message) {
	p.messageQueue <- message
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
		client.Publish(deviceId, 1, false, responseJson)
		return
	}

	if !deviceExists {
		log.Println("connection request from the invalid device, Device Id: ", deviceId)

		response := models.ConnectionUpdateResponse{
			MessageType: 1,
			ErrorStatus: 1,
		}

		responseJson, _ := json.Marshal(response)
		client.Publish(deviceId, 1, false, responseJson)
		return
	}

	if err := p.dbRepo.UpdateDeviceStatus(deviceId, true); err != nil {
		log.Println("error occurred with database while updating the connection status, Device Id: ", deviceId, " Error: ", err.Error())
		response := models.ConnectionUpdateResponse{
			MessageType: 1,
			ErrorStatus: 1,
		}
		responseJson, _ := json.Marshal(response)
		client.Publish(deviceId, 1, false, responseJson)
		return
	}
	response := models.ConnectionUpdateResponse{
		MessageType: 1,
		ErrorStatus: 0,
	}
	responseJson, _ := json.Marshal(response)
	client.Publish(deviceId, 1, false, responseJson)
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
		client.Publish(deviceId, 1, false, responseJson)
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
		client.Publish(deviceId, 1, false, responseJson)
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
		client.Publish(deviceId, 1, false, responseJson)
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
	client.Publish(deviceId, 1, false, responseJson)

}

func (p *messageProcessor) processDeviceDeleteSyncAckRequest(client mqtt.Client, deviceId string, message []byte) {

	req := new(models.DeleteSyncAckRequest)

	if err := json.Unmarshal(message, req); err != nil {
		log.Println("invalid json format in the delete sync ack request, Device Id: ", deviceId, " Error: ", err.Error())

		response := models.DeleteSyncAckResponse{
			MessageType: 3,
			ErrorStatus: 1,
		}

		client.Publish(deviceId, 1, false, response)
		return
	}

	if err := p.dbRepo.DeleteStudentFromDeletes(deviceId, strconv.Itoa(int(req.StudentId))); err != nil {
		log.Println("error occurred with database while deleting the student from deletes, Device Id: ", deviceId, " Error: ", err.Error())
		response := models.DeleteSyncAckResponse{
			MessageType: 3,
			ErrorStatus: 1,
		}
		client.Publish(deviceId, 1, false, response)
		return
	}

	response := models.DeleteSyncAckResponse{
		MessageType: 3,
		ErrorStatus: 0,
	}

	client.Publish(deviceId, 1, false, response)
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
		client.Publish(deviceId, 1, false, responseJson)
		return
	}

	if !exists {
		response := models.InsertSyncResponse{
			MessageType:   4,
			ErrorStatus:   0,
			StudentsEmpty: 1,
		}
		responseJson, _ := json.Marshal(response)
		client.Publish(deviceId, 1, false, responseJson)
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
		client.Publish(deviceId, 1, false, responseJson)
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
	client.Publish(deviceId, 1, false, responseJson)
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
		client.Publish(deviceId, 1, false, responseJson)
		return
	}

	if err := p.dbRepo.DeleteStudentFromInserts(deviceId, strconv.Itoa(int(req.StudentId))); err != nil {
		log.Println("error occurred while deleting the student from inserts, Device Id: ", deviceId, " Error: ", err.Error())

		response := models.InsertSyncAckResponse{
			MessageType: 5,
			ErrorStatus: 1,
		}

		responseJson, _ := json.Marshal(response)
		client.Publish(deviceId, 1, false, responseJson)
		return
	}

	response := models.InsertSyncAckResponse{
		MessageType: 5,
		ErrorStatus: 0,
	}

	responseJson, _ := json.Marshal(response)
	client.Publish(deviceId, 1, false, responseJson)
}

func (p *messageProcessor) processAttendanceRequest(client mqtt.Client, deviceId string, message []byte) {

	req := new(models.UpdateAttendanceRequest)

	if err := json.Unmarshal(message, req); err != nil {
		log.Println("error occurred while decoding the json in update attendance request, DeviceId:", deviceId, " Error: ", err.Error())

		response := models.UpdateAttendanceResponse{
			MessageType: 6,
			ErrorStatus: 1,
		}

		responseJson, _ := json.Marshal(response)
		client.Publish(deviceId, 1, false, responseJson)
		return
	}

	// status, err := p.cacheRepo.CheckMessageDuplication(req.MessageId)

	// if err != nil {
	// 	log.Println("error occurred in redis while checking attendance message duplication, DeviceId: ", deviceId, " Error: ", err.Error())
	// 	response := models.UpdateAttendanceResponse{
	// 		MessageType: 6,
	// 		ErrorStatus: 1,
	// 	}

	// 	responseJson, _ := json.Marshal(response)
	// 	client.Publish(deviceId, 1, false, responseJson)
	// 	return
	// }

	// status := true

	log.Println(req.StudentUnitId, req.TimeStamp)

	if true {
		studentId, err := p.dbRepo.GetStudentId(deviceId, strconv.Itoa(int(req.StudentUnitId)))

		if err != nil {
			log.Println("error occurred while updating the student attendance, DeviceId: ", deviceId, "StudentUnitId: ", req.StudentUnitId, " Error: ", err.Error())
			response := models.UpdateAttendanceResponse{
				MessageType: 6,
				ErrorStatus: 1,
			}

			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}

		t, err := time.Parse("2006-01-02T15:04:05", req.TimeStamp)

		if err != nil {
			log.Println("error occurred while parsing the attendance timestamp, DeviceId: ", deviceId, "StudentUnitId: ", req.StudentUnitId, " Error: ", err.Error())
			response := models.UpdateAttendanceResponse{
				MessageType: 6,
				ErrorStatus: 1,
			}

			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
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
			}

			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}

		if isLogout {
			if err := p.dbRepo.UpdateAttendanceLog(studentId, date, tm); err != nil {
				log.Println("error occurred while updating the student attendance, DeviceId: ", deviceId, "StudentUnitId: ", req.StudentUnitId, " Error: ", err.Error())
				response := models.UpdateAttendanceResponse{
					MessageType: 6,
					ErrorStatus: 1,
				}

				responseJson, _ := json.Marshal(response)
				client.Publish(deviceId, 1, false, responseJson)
				return
			}

			response := models.UpdateAttendanceResponse{
				MessageType: 6,
				ErrorStatus: 0,
				Index:       req.Index,
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
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
				}

				responseJson, _ := json.Marshal(response)
				client.Publish(deviceId, 1, false, responseJson)
				return
			}

			response := models.UpdateAttendanceResponse{
				MessageType: 6,
				ErrorStatus: 0,
				Index:       req.Index,
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
		}

	}

}
