package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"context"
	"time"
	
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main(){
    fmt.Println("Hello Dad Jokes World")
	r := gin.Default()
	r.GET("/ping", GetPing)
	r.GET("/joke", GetJoke)
	r.POST("/joke", CreateJoke)
	r.Run()
}

type CreateJokeBody struct{
	JokeName string `json:"joke_name" bson:"jokeName"`
	JokeContent string `json:"joke_content" bson:"jokeContent"`
}

func GetPing(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func GetJoke(c *gin.Context) {
	param := c.Query("jokeName")
	retrievedJoke, retrieveErr := GetJokeFromDatabase(param)
	if retrieveErr != nil{
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

	createErr := InsertJokeToDatabase(jokeBody)
	if createErr != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": createErr.Error()})
	return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nice joke!"})

}

const mongoUri = "mongodb+srv://DadJokes:dadJ0kes!@cluster0.alp3f.mongodb.net/dadJokes?retryWrites=true&w=majority"

func ConnectToMongo() (collection *mongo.Collection, err error){
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		mongoUri,
	))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			return 
		}
	}()

	err = client.Ping(ctx, nil)

	if err != nil {
		fmt.Println("There was a problem connecting to your Atlas cluster. Check that the URI includes a valid username and password, and that your IP address has been added to the access list. Error: ")
		return nil, err
	}

	var dbName = "dadJokes"
	var collectionName = "jokes"
	collection = client.Database(dbName).Collection(collectionName)
	return collection, nil 
}

func InsertJokeToDatabase(joke CreateJokeBody)error{
	collection, connectErr := ConnectToMongo()
	if connectErr != nil{
		fmt.Println("Error connecting to Mongo", connectErr)
		return connectErr
	}

	jokes := []interface{}{joke}
	insertManyResult, err := collection.InsertMany(context.TODO(), jokes)
	if err != nil {
		fmt.Println("Something went wrong trying to insert the new documents:")
		return err
	}

	fmt.Println("documents successfully inserted.", insertManyResult)

	return nil
}

func  GetJokeFromDatabase(keyword string)(retrievedJoke CreateJokeBody, err error){
	collection, connectErr := ConnectToMongo()
	if connectErr != nil{
		fmt.Println("Error connecting to Mongo", connectErr)
		return CreateJokeBody{}, connectErr
	}

	var result CreateJokeBody
	var myFilter = bson.D{{"jokeName", keyword}}
	e := collection.FindOne(context.TODO(), myFilter).Decode(&result)
	if e != nil {
		fmt.Println("Something went wrong trying to find one document:")
		return CreateJokeBody{}, e
	}
	fmt.Println("Found a document", result)

	return result, nil
}