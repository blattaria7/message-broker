package main

import (
	"flag"
	"log"
	"message-broker/broker"
	"net/http"
)

func main() {
	broker := broker.NewBroker()
	handler := NewHandler(broker)

	http.HandleFunc("/", handler.Handle)

	port := flag.String("port", ":8080", "port to run service")
	flag.Parse()

	log.Printf("message-broker starts on %s port\n", *port)
	if err := http.ListenAndServe(*port, nil); err != nil {
		log.Println(err)
	}
}
