package db

import (
	"context"
	"errors"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"
)

const dbName = "drawConnect"
const collectionName = "users"

//User define the user model from the database
type User struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	Email      string             `json:"email" bson:"email"`
	Signatures []Signature        `json:"signatures" bson:"signatures"`
	Token      string             `json:"token" bson:"token"`
}

//Signature define the signature model received from the phone of the user
type Signature struct {
	Abs  []int `json:"abs" bson:"abs"`
	Ord  []int `json:"ord" bson:"ord"`
	Time []int `json:"time" bson:"time"`
}

//InitDB initialize the mongo client to the host and the port specified
func InitDB(hostName string, dbPort string) (*mongo.Client, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, "mongodb://"+hostName+":"+dbPort)
	if err != nil {
		return nil, err
	}
	return client, nil
}

//PingDBClient return a ping to the client to check if the connection succeeded
func PingDBClient(client *mongo.Client) error {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return client.Ping(ctx, readpref.Primary())
}

//RetrieveUserByID return the user identified by the id or an empty array if none was found
func RetrieveUserByID(client *mongo.Client, id string) ([]User, error) {

	userCollection := client.Database(dbName).Collection(collectionName)
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": objectID}
	return getResultsFind(filter, userCollection)
}

//RetrieveUserByEmail return the user identified by his email or an empty array if none was found
func RetrieveUserByEmail(client *mongo.Client, email string) ([]User, error) {

	userCollection := client.Database(dbName).Collection(collectionName)
	filter := bson.M{"email": email}
	user, err := getResultsFind(filter, userCollection)
	if err != nil {
		return nil, err
	}
	if len(user) > 1 {
		return nil, errors.New("More than one user found")
	}
	return user, nil
}

//RetrieveUsers return all the users in the collection
func RetrieveUsers(client *mongo.Client) ([]User, error) {

	userCollection := client.Database(dbName).Collection(collectionName)
	filter := bson.M{}
	return getResultsFind(filter, userCollection)
}

//InsertUser create a new document in the collection user with the attribute email set to the parameter provided
func InsertUser(client *mongo.Client, email string, signatures []Signature, token string) (string, error) {

	userCollection := client.Database(dbName).Collection(collectionName)
	user := bson.M{"email": email, "signatures": signatures, "token": token}
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

func getResultsFind(filter bson.M, collection *mongo.Collection) ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	results := make([]User, 0)
	for cur.Next(ctx) {
		var result User
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
