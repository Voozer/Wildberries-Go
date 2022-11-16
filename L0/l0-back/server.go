package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nats-io/stan.go"

	_ "github.com/lib/pq"
)

// printing line number on errors
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type order_type map[string]interface{}

var orders_map map[string][]byte
var orders_db *sql.DB

// '/orders/order_uid' handler
func returnOrderJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	order_uid := vars["order_uid"]

	// return from orders map which already has all the data from database retrieved
	if order, isInMap := orders_map[order_uid]; isInMap {
		// fmt.Fprintf(w, "JSON with order_uid = %v...\n\n", order_uid)

		var order_data order_type
		err := json.Unmarshal(order, &order_data)
		if err != nil {
			log.Fatal(err)
		}
		// MAKE IT PRETTIER
		json.NewEncoder(w).Encode(order_data)
	} else {
		// fmt.Fprintf(w, "JSON with order_uid = %v NOT FOUND...\n\n", order_uid)
	}

	fmt.Println("Endpoint: returnOrderJSON")
}

// handling requests
func requestHandler(muxRouter *mux.Router) {
	// order is important!
	// PathPrefix("/") means "/*"
	// so in this case it should be put after more specific path handlers
	muxRouter.HandleFunc("/orders/{order_uid}", returnOrderJSON).Methods("GET")
	muxRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("../l0-front")))
}

// filling in map with db data
func mapFill(orders_db *sql.DB, orders_map map[string][]byte) {
	orders_rows, err := orders_db.Query("select * from orders")
	if err != nil {
		log.Fatal(err)
	}
	defer orders_rows.Close()

	for orders_rows.Next() {
		var (
			order_uid  string
			order_byte []byte
		)
		err := orders_rows.Scan(&order_uid, &order_byte)
		if err != nil {
			log.Fatal(err)
		}
		orders_map[order_uid] = order_byte
	}
}

func main() {
	muxRouter := mux.NewRouter().StrictSlash(true)
	requestHandler(muxRouter)

	// using nats streaming for receiving orders
	// first run the server:
	// go run /home/ilminsky/go/pkg/mod/github.com/nats-io/nats-streaming-server@v0.25.2/nats-streaming-server.go
	sc, err := stan.Connect("test-cluster", "order-sub-id")
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	// using postgresql to store received orders
	// HOW TO STORE PASSWORD?
	connection_params := "user=ilminsky password=ilminskyPSQL dbname=ordersdb"
	orders_db, err = sql.Open("postgres", connection_params)
	if err != nil {
		log.Fatal(err)
	}
	defer orders_db.Close()

	orders_map = make(map[string][]byte)
	mapFill(orders_db, orders_map)

	// receiving orders from order_publisher
	subscription, err := sc.Subscribe("main-channel", func(m *stan.Msg) {
		var order_data order_type
		err = json.Unmarshal(m.Data, &order_data)
		if err != nil {
			log.Fatal(err)
		}

		// adding order to a database
		_, err = orders_db.Exec("insert into orders values ($1, $2) on conflict do nothing", order_data["order_uid"].(string), m.Data)
		if err != nil {
			log.Fatal(err)
		}

		// adding order to an orders map
		orders_map[order_data["order_uid"].(string)] = m.Data

		fmt.Println("Order received")
		fmt.Printf("Received JSON:\n %s\n\n", string(m.Data))

		// fmt.Println()
		// fmt.Println("CURRENT MAP")
		// for id, order_json := range orders_map {
		// 	fmt.Println(id)
		// 	fmt.Println(order_json)
		// }
		// fmt.Println()

	}, stan.DurableName("durable-sub"))
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":8080", muxRouter))

	err = subscription.Close()
	if err != nil {
		log.Fatal(err)
	}

}
