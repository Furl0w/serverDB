package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"
)

var dbPort, dbName, port string

func main() {
	if port = os.Getenv("port"); port == "" {
		port = "3031"
	}
	if dbPort = os.Getenv("dbPort"); dbPort == "" {
		dbPort = "27017"
	}
	if dbName = os.Getenv("dbName"); dbName == "" {
		dbName = "localhost"
	}

	client, err := initDB()
	if err != nil {
		log.Fatal(err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Initiating connection with DB\n")
		err := pingDBClient(client)
		if err != nil {
			fmt.Fprintf(w, "%s\n", err.Error())
			fmt.Fprintf(w, "Failed to connect to DB")
		} else {
			fmt.Fprintf(w, "Succeeded to connect to DB")
		}
		return
	})

	http.HandleFunc("/retrieve", func(w http.ResponseWriter, r *http.Request) {
		results, err := retrieveUserByID(client, "5c3476f7869f6e013359b2fa")
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
		id, err := insertUser(client, "Yo yo")
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

func initDB() (*mongo.Client, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, "mongodb://"+dbName+":"+dbPort)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func pingDBClient(client *mongo.Client) error {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return client.Ping(ctx, readpref.Primary())
}

func retrieveUserByID(client *mongo.Client, id string) ([]primitive.M, error) {

	userCollection := client.Database("drawConnect").Collection("user")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": objectID}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := userCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	results := make([]primitive.M, 0)
	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func insertUser(client *mongo.Client, name string) (string, error) {

	userCollection := client.Database("drawConnect").Collection("user")
	user := bson.M{"name": name}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		return "", err
	}
	if objectid, ok := res.InsertedID.(primitive.ObjectID); ok {
		oid, err := objectid.MarshalJSON()
		if err != nil {
			return "", err
		}
		return string(oid), nil
	}
	return "", errors.New("no id returned")
}
