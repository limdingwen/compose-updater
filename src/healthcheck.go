package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "compose-updater online")
}

func Healthcheck() {
	http.HandleFunc("/", handler)
	log.Println("Starting healthcheck server at port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Println(err)
	}
}
