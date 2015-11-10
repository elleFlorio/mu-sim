package app

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/elleFlorio/testApp/discovery"
	"github.com/elleFlorio/testApp/network"
	"github.com/elleFlorio/testApp/worker"
)

type ServiceParams struct {
	EtcdAddress  string
	Ip           string
	Port         string
	Name         string
	Workload     string
	Destinations []string
}

type Request struct {
	ID      string
	From    string
	Counter int
	Start   time.Time
}

const (
	messagePath  = "/message"
	responsePath = "/response"
)

var (
	name         string
	destinations []string
	workload     string
	requests     map[string]Request
	counter      = 1
	mutex_c      = &sync.Mutex{}
	mutex_m      = &sync.Mutex{}

	ErrNoDestinations = errors.New("No destinations available")
)

func init() {
	requests = make(map[string]Request)
}

func StartService(params ServiceParams) {
	var err error
	name = params.Name
	destinations = params.Destinations
	workload = params.Workload

	err = discovery.InitializeEtcd(params.EtcdAddress)
	if err != nil {
		log.Fatalln("Cannot connect to etcd server at ", params.EtcdAddress)
	}
	log.Println("Connected to etcd server at ", params.EtcdAddress)

	myAddress := network.GenerateAddress(params.Ip, params.Port)
	err = discovery.RegisterToEtcd(params.Name, myAddress)
	if err != nil {
		log.Fatalln("Cannot register to etcd server", params.EtcdAddress)
	}
	log.Println("Registered to etcd server")

	keepAlive()

	http.HandleFunc(responsePath, readResponse)
	http.HandleFunc(messagePath, readMessage)

	log.Println("Waiting for requests...")
	log.Fatal(http.ListenAndServe(params.Port, nil))
}

func keepAlive() {
	go discovery.KeepAlive()
}

func readMessage(w http.ResponseWriter, r *http.Request) {
	var err error
	var requestID string
	var start = time.Now()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	message, err := network.ReadMessage(r)
	if err != nil {
		log.Println("Cannot read message")
		w.WriteHeader(422)
		return
	}
	requestID = message.Args
	if requestID == "" {
		log.Println("New request, generating ID")
		requestID = strconv.Itoa(readAndIncrementCounter())
	}
	log.Printf("Received request %s from %s\n", requestID, message.Sender)

	worker.Work(workload)
	log.Printf("Request %s computed", requestID)
	execTimeMs := time.Since(start).Seconds() * 1000
	log.Printf("Execution time: %fms\n", execTimeMs)

	if len(destinations) > 0 {
		errCounter := sendMessageToDestinations(requestID)
		if errCounter < len(destinations) {
			// This is for requests to multiple destinations
			// because I have to wait till every destination
			// responde me before consider the request complete
			reqCounter := len(destinations) - errCounter
			req := Request{requestID, message.Sender, reqCounter, start}
			addRequestToHistory(req)
		}

		if errCounter > 0 {
			log.Println("Cannot dispatch message to all the destinations")
			w.WriteHeader(422)
			return
		}
	} else {
		respondeToRequest(message.Sender, requestID)
	}

	w.WriteHeader(http.StatusCreated)
}

func sendMessageToDestinations(requestID string) int {
	errCounter := 0

	for _, service := range destinations {
		instances, err := discovery.GetAvailableInstances(service)
		if err != nil {
			log.Println("Cannot dispatch message to service ", service)
			errCounter++
			break
		}
		destination := getDestination(instances)
		sendReqToDest(requestID, destination)
	}

	return errCounter
}

func getDestination(instances []string) string {
	if len(instances) == 1 {
		return instances[0]
	}

	return instances[rand.Intn(len(instances))]
}

func sendReqToDest(reqID string, dest string) {
	go network.Send(dest, "do", reqID, network.GetMyAddress(), false)
	log.Printf("Request %s sent to %s\n", reqID, dest)
}

func readAndIncrementCounter() int {
	mutex_c.Lock()
	c := counter
	counter++
	mutex_c.Unlock()
	runtime.Gosched()

	return c
}

func addRequestToHistory(req Request) {
	mutex_m.Lock()
	requests[req.ID] = req
	mutex_m.Unlock()
	runtime.Gosched()
	log.Printf("Added request %s to history\n", req.ID)
}

func respondeToRequest(dest string, reqId string) {
	network.Send(dest, "done", reqId, network.GetMyAddress(), true)
	log.Printf("Response to request %s sent to %s\n", reqId, dest)
}

func readResponse(w http.ResponseWriter, r *http.Request) {
	var err error
	var respTimeMs float64

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	message, err := network.ReadMessage(r)
	if err != nil {
		log.Println("Cannot read message")
		w.WriteHeader(422)
		return
	}
	log.Println("Received response from ", message.Sender)

	reqId := message.Args
	if req, ok := requests[reqId]; ok {
		respTimeMs = time.Since(req.Start).Seconds() * 1000
		complete := updateRequestInHistory(reqId)
		if complete {
			respondeToRequest(req.From, req.ID)
		}
	} else {
		log.Println("Cannot find request ID in history")
		w.WriteHeader(422)
		return
	}
	log.Printf("Response time: %fms\n", respTimeMs)

	w.WriteHeader(http.StatusCreated)
}

func updateRequestInHistory(reqId string) bool {
	deleted := false
	mutex_m.Lock()
	req := requests[reqId]
	req.Counter -= 1
	if req.Counter == 0 {
		delete(requests, reqId)
		deleted = true
		log.Printf("Removed request %s from history\n", reqId)
	} else {
		requests[reqId] = req
		log.Printf("Updated counter  of request %s: %d\n", reqId, req.Counter)
	}
	mutex_m.Unlock()
	runtime.Gosched()
	return deleted
}
