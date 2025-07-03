package handlers

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/vithsutra/biometric-project-message-processor/models"
)

type messageHandler struct {
	dbRepo    models.DeviceDatabseInterface
	queueRepo models.DeviceQueueInterface
}

func NewMessageHandler(dbRepo models.DeviceDatabseInterface, queueRepo models.DeviceQueueInterface) *messageHandler {
	return &messageHandler{
		dbRepo,
		queueRepo,
	}
}

func (h *messageHandler) DeviceConnectionRequestHandler(client mqtt.Client, message mqtt.Message) {
	deviceIdArr := strings.Split(string(message.Topic()), "/")
	if len(deviceIdArr) > 0 {
		deviceId := deviceIdArr[0]
		deviceExists, err := h.dbRepo.CheckDeviceExists(deviceId)
		if err != nil {
			log.Println("error occurred with database while checking device exists, Device Id: ", deviceId, " Error: ", err.Error())
			response := models.ConnectionUpdateResponse{
				MessageType: "1",
				ErrorStatus: "1",
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}

		if !deviceExists {
			log.Println("connection request from the invalid device, Device Id: ", deviceId)
			response := models.ConnectionUpdateResponse{
				MessageType: "1",
				ErrorStatus: "1",
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}

		if err := h.dbRepo.UpdateDeviceStatus(deviceId, true); err != nil {
			log.Println("error occurred with database while updating the connection status, Device Id: ", deviceId, " Error: ", err.Error())
			response := models.ConnectionUpdateResponse{
				MessageType: "1",
				ErrorStatus: "1",
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}
		response := models.ConnectionUpdateResponse{
			MessageType: "1",
			ErrorStatus: "0",
		}
		responseJson, _ := json.Marshal(response)
		client.Publish(deviceId, 1, false, responseJson)

	} else {
		log.Println("invalid connection request topic , Topic: ", message.Topic())
	}
}

func (h *messageHandler) DeviceDisconnectionRequestHandler(client mqtt.Client, message mqtt.Message) {
	deviceIdArr := strings.Split(string(message.Topic()), "/")
	if len(deviceIdArr) > 0 {
		deviceId := deviceIdArr[0]
		if err := h.dbRepo.UpdateDeviceStatus(deviceId, false); err != nil {
			log.Println("error occurred with database while updating the disconnection status, Device Id: ", deviceId, " Error: ", err.Error())
		}
	} else {
		log.Println("invalid disconnection request topic , Topic: ", message.Topic())
	}
}

func (h *messageHandler) DeleteSyncRequestHandler(client mqtt.Client, message mqtt.Message) {
	deviceIdArr := strings.Split(string(message.Topic()), "/")
	if len(deviceIdArr) > 0 {
		deviceId := deviceIdArr[0]

		exists, err := h.dbRepo.CheckStudentsExistsInDeletes(deviceId)

		if err != nil {
			log.Println("error occurred with database while checking students exists in deletes, Device Id: ", deviceId, " Error: ", err.Error())
			response := models.DeleteSyncResponse{
				MessageType:   "2",
				ErrorStatus:   "1",
				StudentsEmpty: "0",
				StudentId:     0,
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}

		if !exists {
			response := models.DeleteSyncResponse{
				MessageType:   "2",
				ErrorStatus:   "0",
				StudentsEmpty: "1",
				StudentId:     0,
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}

		studentId, err := h.dbRepo.GetStudentFromDeletes(deviceId)

		if err != nil {
			log.Println("error occurred with database while getting student from deletes, Device Id: ", deviceId, " Error: ", err.Error())
			response := models.DeleteSyncResponse{
				MessageType:   "2",
				ErrorStatus:   "1",
				StudentsEmpty: "0",
				StudentId:     0,
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}
		studentIdInt, _ := strconv.Atoi(studentId)

		response := models.DeleteSyncResponse{
			MessageType:   "2",
			ErrorStatus:   "0",
			StudentsEmpty: "0",
			StudentId:     uint32(studentIdInt),
		}
		responseJson, _ := json.Marshal(response)
		client.Publish(deviceId, 1, false, responseJson)
	} else {
		log.Println("invalid disconnection request topic , Topic: ", message.Topic())
	}
}

func (h *messageHandler) DeleteSyncAckRequestHandler(client mqtt.Client, message mqtt.Message) {
	deviceIdArr := strings.Split(string(message.Topic()), "/")
	if len(deviceIdArr) > 0 {
		deviceId := deviceIdArr[0]
		req := new(models.DeleteSyncAckRequest)

		if err := json.Unmarshal(message.Payload(), req); err != nil {
			log.Println("invalid json format in the delete sync ack request, Device Id: ", deviceId, " Error: ", err.Error())
			response := models.DeleteSyncAckResponse{
				MessageType: "3",
				ErrorStatus: "1",
			}
			client.Publish(deviceId, 1, false, response)
			return
		}

		if err := h.dbRepo.DeleteStudentFromDeletes(deviceId, req.StudentId); err != nil {
			log.Println("error occurred with database while deleting the student from deletes, Device Id: ", deviceId, " Error: ", err.Error())
			response := models.DeleteSyncAckResponse{
				MessageType: "3",
				ErrorStatus: "1",
			}
			client.Publish(deviceId, 1, false, response)
			return
		}

		response := models.DeleteSyncAckResponse{
			MessageType: "3",
			ErrorStatus: "0",
		}
		client.Publish(deviceId, 1, false, response)

	} else {
		log.Println("invalid disconnection request topic , Topic: ", message.Topic())
	}
}

func (h *messageHandler) InsertSyncRequestHandler(client mqtt.Client, message mqtt.Message) {
	deviceIdArr := strings.Split(string(message.Topic()), "/")
	if len(deviceIdArr) > 0 {
		deviceId := deviceIdArr[0]
		exists, err := h.dbRepo.CheckStudentsExistsInInserts(deviceId)
		if err != nil {
			log.Println("error occurred with database while checking student exists in inserts, Device Id: ", deviceId, " Error: ", err.Error())
			response := models.InsertSyncResponse{
				MessageType: "4",
				ErrorStatus: "1",
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}

		if !exists {
			response := models.InsertSyncResponse{
				MessageType:   "4",
				ErrorStatus:   "0",
				StudentsEmpty: "1",
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}

		studentId, fingerprintData, err := h.dbRepo.GetStudentFromInserts(deviceId)
		if err != nil {
			log.Println("error occurred with database while getting student from inserts, Device Id: ", deviceId, " Error: ", err.Error())
			response := models.InsertSyncResponse{
				MessageType: "4",
				ErrorStatus: "1",
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}

		studentIdInt, _ := strconv.Atoi(studentId)

		response := models.InsertSyncResponse{
			MessageType:     "4",
			ErrorStatus:     "0",
			StudentsEmpty:   "0",
			StudentId:       uint32(studentIdInt),
			FingerPrintData: fingerprintData,
		}
		responseJson, _ := json.Marshal(response)
		client.Publish(deviceId, 1, false, responseJson)
	} else {
		log.Println("invalid disconnection request topic , Topic: ", message.Topic())
	}
}

func (h *messageHandler) InsertSyncAckRequestHandler(client mqtt.Client, message mqtt.Message) {
	deviceIdArr := strings.Split(string(message.Topic()), "/")
	if len(deviceIdArr) > 0 {
		deviceId := deviceIdArr[0]
		req := new(models.InsertSyncAckRequest)
		if err := json.Unmarshal(message.Payload(), req); err != nil {
			log.Println("error occurred while decoding json insert sync ack message, Device Id: ", deviceId, " Error: ", err.Error())
			response := models.InsertSyncAckResponse{
				MessageType: "5",
				ErrorStatus: "1",
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}

		if err := h.dbRepo.DeleteStudentFromInserts(deviceId, strconv.Itoa(int(req.StudentId))); err != nil {
			log.Println("error occurred while deleting the student from inserts, Device Id: ", deviceId, " Error: ", err.Error())
			response := models.InsertSyncAckResponse{
				MessageType: "5",
				ErrorStatus: "1",
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}
		response := models.InsertSyncAckResponse{
			MessageType: "5",
			ErrorStatus: "0",
		}
		responseJson, _ := json.Marshal(response)
		client.Publish(deviceId, 1, false, responseJson)
	} else {
		log.Println("invalid disconnection request topic , Topic: ", message.Topic())
	}
}

func (h *messageHandler) UpdateAttendanceRequestHandler(client mqtt.Client, message mqtt.Message) {
	deviceIdArr := strings.Split(string(message.Topic()), "/")
	if len(deviceIdArr) > 0 {
		deviceId := deviceIdArr[0]
		req := new(models.UpdateAttendanceRequest)
		if err := json.Unmarshal(message.Payload(), req); err != nil {
			log.Println("error occurred while decoding the json in update attendance request, Device Id:", deviceId, " Error: ", err.Error())
			response := models.UpdateAttendanceResponse{
				MessageType: "6",
				ErrorStatus: "1",
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}

		attendance := models.Attendance{
			DeviceId:  deviceId,
			StudentId: strconv.Itoa(int(req.StudentId)),
			TimeStamp: req.TimeStamp,
		}

		if err := h.queueRepo.PublishAttendance(&attendance); err != nil {
			log.Println("error occurred while publishing the attendance to kafka, Device Id: ", deviceId, " Error: ", err.Error())
			response := models.UpdateAttendanceResponse{
				MessageType: "6",
				ErrorStatus: "1",
			}
			responseJson, _ := json.Marshal(response)
			client.Publish(deviceId, 1, false, responseJson)
			return
		}

		response := models.UpdateAttendanceResponse{
			MessageType: "6",
			Index:       req.Index,
			ErrorStatus: "0",
		}
		responseJson, _ := json.Marshal(response)
		client.Publish(deviceId, 1, false, responseJson)
	} else {
		log.Println("invalid disconnection request topic , Topic: ", message.Topic())
	}
}
