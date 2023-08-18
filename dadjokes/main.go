package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	uuid "github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("Hello Dad Jokes World")
	router := RouterBootstrap()
	routeChan := make(chan string)
	go func() {
		err := router.Run()
		if err != nil {
			routeChan <- err.Error()
		}
	}()

	select {
	case routeErr := <-routeChan:
		fmt.Println("Error starting app:", routeErr)
		os.Exit(1)
	}
}

func RouterBootstrap() *gin.Engine {
	os.Setenv("PORT", "8082")
	r := gin.New()
	//r := gin.Default()
	r.GET("/ping", GetPing)
	r.GET("/joke", GetJoke)
	r.POST("/joke", CreateJoke)
	//r.Run()
	return r
}

type CreateJokeBody struct {
	JokeName    string `json:"joke_name" bson:"jokeName"`
	JokeContent string `json:"joke_content" bson:"jokeContent"`
}

func GetPing(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func GetJoke(c *gin.Context) {
	param := c.Query("jokeName")
	newUUID := uuid.New().String()
	ctx := context.WithValue(c.Request.Context(), "UUID", newUUID)
	retrievedJoke, retrieveErr := GetJokeFromDatabase(ctx, param)
	if retrieveErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": retrieveErr.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": retrievedJoke,
	})
}

func CreateJoke(c *gin.Context) {
	var jokeBody CreateJokeBody
	if err := c.ShouldBindJSON(&jokeBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newUUID := uuid.New().String()
	ctx := context.WithValue(c.Request.Context(), "UUID", newUUID)

	createErr := InsertJokeToDatabase(ctx, jokeBody)
	if createErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": createErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nice joke!"})

}

const mongoUri = "mongodb://root:example@mongo:27017/"

func ConnectToMongo(ctx context.Context) (collection *mongo.Collection, err error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		mongoUri,
	))

	err = client.Ping(ctx, nil)

	if err != nil {
		fmt.Println("Error connecting to mongo: ", err)
		return nil, err
	}

	var dbName = "dadJokes"
	var collectionName = "jokes"
	collection = client.Database(dbName).Collection(collectionName)
	return collection, nil
}

type MongoJoke struct {
	ID          primitive.ObjectID `bson:"id,omitempty`
	JokeName    string             `bson:"joke_name,omitempty`
	JokeContent string             `bson:"joke_content,omitempty`
}

func InsertJokeToDatabase(ctx context.Context, joke CreateJokeBody) error {
	collection, connectErr := ConnectToMongo(ctx)
	if connectErr != nil {
		fmt.Println("Error connecting to Mongo", connectErr)
		return connectErr
	}

	jokes := MongoJoke{
		JokeName:    joke.JokeName,
		JokeContent: joke.JokeContent,
	}
	insertManyResult, err := collection.InsertOne(context.TODO(), jokes)
	if err != nil {
		fmt.Println("Something went wrong trying to insert the new documents: ", err)
		return err
	}

	fmt.Println("documents successfully inserted", insertManyResult)
	fmt.Println("insert ID", insertManyResult.InsertedID)

	return nil
}

func GetJokeFromDatabase(ctx context.Context, keyword string) (retrievedJoke CreateJokeBody, err error) {
	collection, connectErr := ConnectToMongo(ctx)
	if connectErr != nil {
		fmt.Println("Error connecting to Mongo", connectErr)
		return CreateJokeBody{}, connectErr
	}

	var result MongoJoke
	var myFilter = bson.D{{Key: "joke_name", Value: keyword}}
	findErr := collection.FindOne(ctx, myFilter).Decode(&result)
	if findErr != nil {
		fmt.Println("Something went wrong trying to find one document: ", findErr)
		return CreateJokeBody{}, findErr
	}

	fmt.Println("Found a document", result)
	jokeBody := CreateJokeBody{
		JokeName:    result.JokeName,
		JokeContent: result.JokeContent,
	}
	return jokeBody, nil
}
