package app

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/elleFlorio/testApp/discovery"
	"github.com/elleFlorio/testApp/metric"
	"github.com/elleFlorio/testApp/network"
	"github.com/elleFlorio/testApp/worker"
)

type ServiceParams struct {
	EtcdAddress   string
	InfluxAddress string
	InfluxDbName  string
	InfluxUser    string
	InfluxPwd     string
	Ip            string
	Port          string
	Name          string
	Workload      string
	Destinations  []string
}

const (
	messagePath  = "/message"
	responsePath = "/response"
)

var (
	name         string
	destinations []string
	workload     string
	requests     map[string]network.Request
	jobs         map[string]network.Request
	counter      = 1
	mutex_c      = &sync.Mutex{}
	mutex_r      = &sync.Mutex{}
	mutex_w      = &sync.Mutex{}
	ch_req       chan network.Request
	ch_stop      chan struct{}

	ErrNoDestinations = errors.New("No destinations available")
)

func init() {
	requests = make(map[string]network.Request)
	jobs = make(map[string]network.Request)
	ch_req = make(chan network.Request)
	ch_stop = make(chan struct{})
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

	startSigsMonitor(ch_stop)
	keepAlive(ch_stop)
	startJobsManager(ch_req)
	initializeMetricService(params)

	http.HandleFunc(responsePath, readResponse)
	http.HandleFunc(messagePath, readMessage)

	log.Println("Waiting for requests...")
	log.Fatal(http.ListenAndServe(params.Port, nil))
}

func keepAlive(ch_stop chan struct{}) {
	go discovery.KeepAlive(ch_stop)
}

func startJobsManager(ch_req chan network.Request) {
	go jobsManager(ch_req)
}

func jobsManager(ch_req chan network.Request) {
	log.Println("Started work manager. Waiting for work to do...")
	ch_done := make(chan network.Request)
	for {
		select {
		case req := <-ch_req:
			log.Println("Starting new worker on request ", req.ID)
			addReqToWorks(req)
			go worker.Work(workload, req, ch_done)
		case reqDone := <-ch_done:
			log.Printf("Request %s computed", reqDone.ID)
			log.Println("service " + name + " " + "execution_time:" + strconv.FormatFloat(reqDone.ExecTimeMs, 'f', 2, 64) + "ms")
			finalizeReq(reqDone)
			removeReqFromWorks(reqDone.ID)
			metric.SendExecutionTime(reqDone.ExecTimeMs)
		}
	}
}

func startSigsMonitor(ch_stop chan struct{}) {
	go sigsMonitor(ch_stop)
}

func sigsMonitor(ch_stop chan struct{}) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigs:
			go shutDown(ch_stop)
		}
	}
}

func shutDown(ch_stop chan struct{}) {
	log.Println("Received shutdown signal")
	ch_stop <- struct{}{}
	log.Println("Stopped keep alive goroutine")
	discovery.UnregisterFromEtcd()
	log.Println("Unregistered from etcd")
	for isServiceWorking() {
		log.Println("Waiting for jobs to complete...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	for isServiceWaiting() {
		log.Println("Waiting for responses to requests...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Fatalln("Done. Shutting down")
}

func isServiceWorking() bool {
	mutex_w.Lock()
	jobsInProgress := len(jobs)
	mutex_w.Unlock()

	if jobsInProgress != 0 {
		return true
	}

	return false
}

func isServiceWaiting() bool {
	mutex_r.Lock()
	requestsPending := len(requests)
	mutex_r.Unlock()

	if requestsPending != 0 {
		return true
	}

	return false
}

func initializeMetricService(params ServiceParams) {
	config := metric.InfluxConfig{
		params.InfluxAddress,
		params.InfluxDbName,
		params.InfluxUser,
		params.InfluxPwd,
	}
	err := metric.Initialize(params.Name, params.Workload, params.Ip, config)
	if err != nil {
		log.Fatalf("Error: %s; failded to initialize metric service", err.Error())
	}
}

func addReqToWorks(req network.Request) {
	mutex_w.Lock()
	jobs[req.ID] = req
	mutex_w.Unlock()
}

func removeReqFromWorks(id string) {
	mutex_w.Lock()
	delete(jobs, id)
	mutex_w.Unlock()
}

func finalizeReq(reqDone network.Request) {
	if reqDone.To != "" {
		err := sendMessageToSpecificService(reqDone.ID, reqDone.To)
		if err != nil {
			log.Println("Cannot dispatch message to service", reqDone.To)
			return
		}
		addRequestToHistory(reqDone)
	} else {
		if len(destinations) > 0 {
			errCounter := sendMessageToDestinations(reqDone.ID)
			if errCounter < len(destinations) {
				// This is for requests to multiple destinations
				// because I have to wait till every destination
				// responde me before consider the request complete
				reqCounter := len(destinations) - errCounter
				reqDone.Counter = reqCounter
				addRequestToHistory(reqDone)
			}
			if errCounter > 0 {
				log.Println("Cannot dispatch message to all the destinations")
				if errCounter == len(destinations) {
					respondeToRequest(reqDone.From, reqDone.ID, "done")
				}
				return
			}
		} else {
			respondeToRequest(reqDone.From, reqDone.ID, "done")
		}
	}
}

func readMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Create the request
	req, err := createReq(r)
	if err != nil {
		log.Println("Cannot read message")
		w.WriteHeader(422)
		return
	}

	// Start work
	ch_req <- req

	w.WriteHeader(http.StatusCreated)
}

func createReq(r *http.Request) (network.Request, error) {
	var err error
	var requestID string
	var start = time.Now()

	//read request
	message, err := network.ReadMessage(r)
	if err != nil {
		log.Println("Cannot read message")
		return network.Request{}, err
	}
	requestID = message.Args
	if requestID == "" {
		log.Println("New request, generating ID")
		requestID = strconv.Itoa(readAndIncrementCounter())
	}
	log.Printf("Received request %s from %s\n", requestID, message.Sender)

	toService, err := network.ReadParam("service", r)
	if err != nil {
		log.Println("Cannot read param 'service'")
	}

	req := network.Request{
		ID:         requestID,
		From:       message.Sender,
		To:         toService,
		Counter:    len(destinations),
		Start:      start,
		ExecTimeMs: 0,
	}

	return req, nil
}

func sendMessageToSpecificService(requestID string, service string) error {
	instances, err := discovery.GetAvailableInstances(service)
	if err != nil {
		log.Println("Cannot dispatch message to service ", service)
		return err
	}
	destination := getDestination(instances)
	sendReqToDest(requestID, destination)
	return nil
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

func addRequestToHistory(req network.Request) {
	mutex_r.Lock()
	requests[req.ID] = req
	mutex_r.Unlock()
	runtime.Gosched()
	log.Printf("Added request %s to history\n", req.ID)
}

func respondeToRequest(dest string, reqId string, status string) {
	network.Send(dest, status, reqId, network.GetMyAddress(), true)
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
			respondeToRequest(req.From, req.ID, message.Body)
		}
	} else {
		log.Println("Cannot find request ID in history")
		w.WriteHeader(422)
		return
	}
	if message.Body == "done" {
		log.Println("service " + name + " " + "response_time" + ":" + strconv.FormatFloat(respTimeMs, 'f', 2, 64) + "ms")
		metric.SendResponseTime(respTimeMs)
	} else {
		log.Println("Error: request lost.")
	}

	w.WriteHeader(http.StatusCreated)
}

func updateRequestInHistory(reqId string) bool {
	deleted := false
	mutex_r.Lock()
	req := requests[reqId]
	req.Counter -= 1
	if req.Counter <= 0 {
		delete(requests, reqId)
		deleted = true
		log.Printf("Removed request %s from history\n", reqId)
	} else {
		requests[reqId] = req
		log.Printf("Updated counter  of request %s: %d\n", reqId, req.Counter)
	}
	mutex_r.Unlock()
	runtime.Gosched()
	return deleted
}
