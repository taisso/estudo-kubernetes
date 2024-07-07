package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var startedAt = time.Now()

func main() {

	http.HandleFunc("/healthz", Healthz)
	http.HandleFunc("/config-map", ConfigMap)
	http.HandleFunc("/secret", Secret)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		name := os.Getenv("NAME")
		age := os.Getenv("AGE")
		fmt.Fprintf(w, "Hello, I'm %s. I'm %s.", name, age)
	})

	http.ListenAndServe(":8080", nil)
}

func ConfigMap(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("/go/myfamily/family.txt")
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	fmt.Fprintf(w, "My Family: %s\n", string(data))
}

func Secret(w http.ResponseWriter, r *http.Request) {
	user := os.Getenv("USER")
	password := os.Getenv("PASSWORD")

	fmt.Fprintf(w, "User: %s. Password: %s", user, password)
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	duration := time.Since(startedAt)

	if duration.Seconds() < 10 || duration.Seconds() > 30 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Duration: %v", duration.Seconds())))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
