package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type Response struct {
	Message string `json:"message"`
}

func main() {

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
