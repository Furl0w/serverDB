package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"serverMongoDB/db"
	"strings"
	"unicode"

	"github.com/gorilla/mux"
	"github.com/mongodb/mongo-go-driver/mongo"
)

var client *mongo.Client

type creationRequest struct {
	Name string `json:"name,omitempty"`
}

type successInsert struct {
	ID string `json:"id,omitempty"`
}

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

	c, err := db.InitDB(hostName, dbPort)
	if err != nil {
		log.Fatal(err)
		return
	}
	client = c

	r := mux.NewRouter()
	r.HandleFunc("/", pingDB).Methods("GET")
	r.HandleFunc("/users", getUsers).Methods("GET")
	r.HandleFunc("/user/{id}", getUser).Methods("GET")
	r.HandleFunc("/user", createUser).Methods("POST")

	log.Fatal(http.ListenAndServe(":"+port, r))
}

func pingDB(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Initiating connection with DB\n")
	err := db.PingDBClient(client)
	if err != nil {
		w.WriteHeader(500)
		log.Print(err.Error())
		fmt.Fprintf(w, "Failed to connect to DB")
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Succeeded to connect to DB")
	return
}

func getUsers(w http.ResponseWriter, r *http.Request) {

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
		log.Print(err.Error())
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", jsonString)
	return
}

func getUser(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	results, err := db.RetrieveUserByID(client, params["id"])
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "error when finding stuff\n")
		log.Print(err.Error())
		return
	}
	jsonString, err := json.Marshal(results)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "error when marshalling json\n")
		log.Print(err.Error())
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", jsonString)
	return
}

func createUser(w http.ResponseWriter, r *http.Request) {

	var userRequest creationRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "error when receiving stuff\n")
		log.Print(err.Error())
		return
	}
	err = json.Unmarshal(body, &userRequest)
	if err != nil {
		w.WriteHeader(403)
		fmt.Fprintf(w, "error when unmarshalling json\n")
		log.Print(err.Error())
		return
	}
	id, err := db.InsertUser(client, userRequest.Name)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "error when inserting stuff\n")
		log.Print(err.Error())
		return
	}
	w.Header().Set("Content-type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	id = strings.TrimFunc(id, func(r rune) bool {
		return r == '"' || unicode.IsSpace(r)
	})
	res := successInsert{ID: id}
	json, err := json.Marshal(res)
	fmt.Fprint(w, string(json))
	return
}
