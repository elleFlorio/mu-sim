package network

import (
	"bytes"
	"encoding/json"
	"log"
	//"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Message struct {
	Sender string `json:"sender"`
	Body   string `json:"body"`
	Args   string `json:"args"`
}

func doRequest(method string, path string, body []byte) {
	b := bytes.NewBuffer(body)
	req, err := http.NewRequest(method, path, b)

	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-type", "application/json")

	client := &http.Client{}
	client.Do(req)
}

func Send(address string, message string, args string, from string, isResponse bool) {
	path := address + "/message"
	if isResponse {
		path = address + "/response"
	}
	m := Message{
		Sender: from,
		Body:   message,
		Args:   args,
	}

	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	doRequest("POST", path, data)
}

func ReadMessage(r *http.Request) (Message, error) {
	var err error
	var message Message

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Println(err)
		return Message{}, err
	}

	if err = r.Body.Close(); err != nil {
		log.Println(err)
		return Message{}, err
	}

	if err = json.Unmarshal(body, &message); err != nil {
		log.Println(err)
		return Message{}, err
	}

	return message, nil
}
