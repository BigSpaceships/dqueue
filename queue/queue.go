package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"

	"github.com/bigspaceships/circlejerk/auth"
	dq_websocket "github.com/bigspaceships/circlejerk/websocket"
)

type QueueEntry struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Type     string `json:"type"`
	Id       int    `json:"id"`
}

type QueueRequestData struct {
	Type string `json:"type"`
}

type Queue struct {
	Id         int          `json:"id"`
	Points     []QueueEntry `json:"points"`
	Clarifiers []QueueEntry `json:"clarifiers"`
	Children   []Queue      `json:"children"`
	Topic      string       `json:"topic"`
	pointCount int
}

type Discussion struct {
	Queue      Queue `json:"queue"`
	queueMap   map[int]*Queue
	wsServer   *dq_websocket.WsServer
	queueCount int
}

func SetupDiscussion(wsServer *dq_websocket.WsServer) *Discussion {
	discussion := Discussion{
		wsServer:   wsServer,
		queueCount: 0,
		queueMap:   make(map[int]*Queue),
	}

	baseQueue := Queue{
		Id:         discussion.queueCount,
		Points:     make([]QueueEntry, 0),
		Clarifiers: make([]QueueEntry, 0),
		Children:   make([]Queue, 0),
		Topic:      "Big long discussion",
		pointCount: 0,
	}

	discussion.queueCount++
	discussion.Queue = baseQueue
	discussion.queueMap[baseQueue.Id] = &baseQueue

	return &discussion
}

func (discussion *Discussion) DeletePoint(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserClaims(r)

	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, "error parsing id to int"+err.Error(), http.StatusBadRequest)
		return
	}

	queue_id, err := strconv.Atoi(r.PathValue("queue"))

	if err != nil {
		http.Error(w, "error parsing queue_id to int"+err.Error(), http.StatusBadRequest)
		return
	}

	queue, ok := discussion.queueMap[queue_id]

	if !ok {
		http.Error(w, fmt.Sprintf("queue %d was not a part of the discussion", queue_id), http.StatusBadRequest)
	}

	pointIndex := slices.IndexFunc(queue.Points, func(entry QueueEntry) bool {
		return entry.Id == id
	})

	point := queue.Points[pointIndex]

	if !(userInfo.IsEboard || point.Username == userInfo.Username) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	queue.Points = slices.Concat(queue.Points[:pointIndex], queue.Points[(pointIndex+1):])
	if queue.Points == nil {
		queue.Points = make([]QueueEntry, 0)
	}

	discussion.wsServer.SendWSMessage(struct {
		Type      string `json:"type"`
		Id        int    `json:"id"`
		Dismisser string `json:"dismisser"`
	}{
		Type:      "delete",
		Id:        id,
		Dismisser: userInfo.Name,
	})
}

func (discussion *Discussion) DeleteClarifier(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserClaims(r)

	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, "error parsing id to int"+err.Error(), http.StatusBadRequest)
		return
	}

	queue_id, err := strconv.Atoi(r.PathValue("queue"))

	if err != nil {
		http.Error(w, "error parsing queue_id to int"+err.Error(), http.StatusBadRequest)
		return
	}

	queue, ok := discussion.queueMap[queue_id]

	if !ok {
		http.Error(w, fmt.Sprintf("queue %d was not a part of the discussion", queue_id), http.StatusBadRequest)
	}

	pointIndex := slices.IndexFunc(queue.Clarifiers, func(entry QueueEntry) bool {
		return entry.Id == id
	})

	point := queue.Clarifiers[pointIndex]

	if !(userInfo.IsEboard || point.Username == userInfo.Username) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	queue.Clarifiers = slices.Concat(queue.Clarifiers[:pointIndex], queue.Clarifiers[(pointIndex+1):])
	if queue.Clarifiers == nil {
		queue.Clarifiers = make([]QueueEntry, 0)
	}

	discussion.wsServer.SendWSMessage(struct {
		Type      string `json:"type"`
		Id        int    `json:"id"`
		Dismisser string `json:"dismisser"`
	}{
		Type:      "delete",
		Id:        id,
		Dismisser: userInfo.Name,
	})
}

func (discussion *Discussion) NewClarifier(w http.ResponseWriter, r *http.Request) {
	log.Println("hi")
	userInfo := auth.GetUserClaims(r)
	log.Println("hi")

	queue_id, err := strconv.Atoi(r.PathValue("queue"))

	log.Println("hi")
	if err != nil {
		http.Error(w, "error parsing queue_id to int"+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("%d\n", queue_id)

	queue, ok := discussion.queueMap[queue_id]

	log.Println("hi")
	if !ok {
		http.Error(w, fmt.Sprintf("queue %d was not a part of the discussion", queue_id), http.StatusBadRequest)
	}

	requestData := QueueRequestData{}
	json.NewDecoder(r.Body).Decode(&requestData)

	queue.pointCount++

	newEntry := QueueEntry{
		Name:     userInfo.Name,
		Username: userInfo.Username,
		Type:     "clarifier",
		Id:       queue.pointCount,
	}

	queue.Clarifiers = append(queue.Clarifiers, newEntry)

	w.WriteHeader(http.StatusOK)

	discussion.wsServer.SendWSMessage(struct {
		Type string     `json:"type"`
		Data QueueEntry `json:"data"`
	}{
		Type: "clarifier",
		Data: newEntry,
	})
}

func (discussion *Discussion) NewPoint(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserClaims(r)

	queue_id, err := strconv.Atoi(r.PathValue("queue"))

	if err != nil {
		http.Error(w, "error parsing queue_id to int"+err.Error(), http.StatusBadRequest)
		return
	}

	queue, ok := discussion.queueMap[queue_id]

	if !ok {
		http.Error(w, fmt.Sprintf("queue %d was not a part of the discussion", queue_id), http.StatusBadRequest)
	}

	requestData := QueueRequestData{}
	json.NewDecoder(r.Body).Decode(&requestData)

	queue.pointCount++

	newEntry := QueueEntry{
		Name:     userInfo.Name,
		Username: userInfo.Username,
		Type:     "point",
		Id:       queue.pointCount,
	}

	queue.Points = append(queue.Points, newEntry)

	w.WriteHeader(http.StatusOK)

	discussion.wsServer.SendWSMessage(struct {
		Type string     `json:"type"`
		Data QueueEntry `json:"data"`
	}{
		Type: "point",
		Data: newEntry,
	})
}

func (discussion *Discussion) ChangeTopic(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserClaims(r)

	if !userInfo.IsEboard {
		http.Error(w, "user is not on eboard", http.StatusForbidden)
		return
	}

	queue_id, err := strconv.Atoi(r.PathValue("queue"))

	if err != nil {
		http.Error(w, "error parsing queue_id to int"+err.Error(), http.StatusBadRequest)
		return
	}

	queue, ok := discussion.queueMap[queue_id]

	if !ok {
		http.Error(w, fmt.Sprintf("queue %d was not a part of the discussion", queue_id), http.StatusBadRequest)
	}

	requestData := struct {
		NewTopic string `json:"new-topic"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&requestData)

	if err != nil {
		http.Error(w, "Error decoding body:"+err.Error(), http.StatusBadRequest)
		return
	}

	queue.Topic = requestData.NewTopic

	discussion.wsServer.SendWSMessage(struct {
		Type  string `json:"type"`
		Topic string `json:"topic"`
	}{
		Type:  "topic",
		Topic: requestData.NewTopic,
	})

	w.WriteHeader(http.StatusOK)
}

func (discussion *Discussion) GetDiscussion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(discussion)
}

func (discussion *Discussion) GetQueue(w http.ResponseWriter, r *http.Request) {
	queue_id, err := strconv.Atoi(r.PathValue("queue"))

	if err != nil {
		http.Error(w, "error parsing queue_id to int"+err.Error(), http.StatusBadRequest)
		return
	}

	queue, ok := discussion.queueMap[queue_id]

	if !ok {
		http.Error(w, fmt.Sprintf("queue %d was not a part of the discussion", queue_id), http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(queue)
}
