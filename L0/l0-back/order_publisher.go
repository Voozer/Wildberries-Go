package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/nats-io/stan.go"
)

// print line number on errors
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	// using nats streaming
	// first run the server:
	// go run /home/ilminsky/go/pkg/mod/github.com/nats-io/nats-streaming-server@v0.25.2/nats-streaming-server.go

	sc, err := stan.Connect("test-cluster", "order-pub-id")
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	// opening model.json
	model_JSON, err := os.Open("model.json")
	if err != nil {
		log.Fatal(err)
	}
	defer model_JSON.Close()

	model_byte, _ := ioutil.ReadAll(model_JSON)
	var model_data map[string]interface{}
	err = json.Unmarshal(model_byte, &model_data)
	if err != nil {
		log.Fatal(err)
	}

	// modifying order_uid to add new tests
	var i int
	for {
		model_data["order_uid"] = "order" + strconv.Itoa(i)
		order_byte, err := json.Marshal(model_data)
		if err != nil {
			log.Fatal(err)
		}

		// publishing new order to subscriber
		err = sc.Publish("main-channel", order_byte)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Order %v sent\n", i)
		i++

		time.Sleep(5 * time.Second)
	}
}
