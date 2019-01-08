package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"serverMongoDB/db"
)

func main() {
	var dbPort, hostName, port string
	if port = os.Getenv("port"); port == "" {
		port = "3031"
	}
	if dbPort = os.Getenv("dbPort"); dbPort == "" {
		dbPort = "27017"
	}
	if hostName = os.Getenv("hostName"); hostName == "" {
		hostName = "localhost"
	}

	client, err := db.InitDB(hostName, dbPort)
	if err != nil {
		log.Fatal(err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Initiating connection with DB\n")
		err := db.PingDBClient(client)
		if err != nil {
			fmt.Fprintf(w, "%s\n", err.Error())
			fmt.Fprintf(w, "Failed to connect to DB")
		} else {
			fmt.Fprintf(w, "Succeeded to connect to DB")
		}
		return
	})

	http.HandleFunc("/retrieve", func(w http.ResponseWriter, r *http.Request) {
		results, err := db.RetrieveUserByID(client, "5c34835d30195d52cb03b7c2")
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "error when finding stuff\n")
			fmt.Fprintf(w, "%s\n", err.Error())
			return
		}
		jsonString, err := json.Marshal(results)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "error when marshalling json\n")
			fmt.Fprintf(w, "%s\n", err.Error())
			return
		}
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, "%s", jsonString)
		return
	})

	http.HandleFunc("/retrieves", func(w http.ResponseWriter, r *http.Request) {
		results, err := db.RetrieveUsers(client)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "error when finding stuff\n")
			fmt.Fprintf(w, "%s\n", err.Error())
			return
		}
		jsonString, err := json.Marshal(results)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "error when marshalling json\n")
			fmt.Fprintf(w, "%s\n", err.Error())
			return
		}
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, "%s", jsonString)
		return
	})

	http.HandleFunc("/insert", func(w http.ResponseWriter, r *http.Request) {
		id, err := db.InsertUser(client, "Yo yo")
		if err != nil {
			fmt.Fprintf(w, "error when finding stuff\n")
			fmt.Fprintf(w, "%s\n", err.Error())
			return
		}
		fmt.Fprint(w, id)
		return
	})

	http.ListenAndServe(":"+port, nil)
}
