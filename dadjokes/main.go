package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"

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

	r.GET("/ping", GetPing)
	r.GET("/joke", GetJoke)
	r.POST("/joke", CreateJoke)
	r.GET("/jokes/all", GetAllJokes)
	r.GET("/random/joke", GetRandom)

	return r
}

type MongoJoke struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	JokeName    string             `bson:"joke_name" json:"joke_name"`
	JokeContent string             `bson:"joke_content" json:"joke_content"`
}

func GetPing(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func GetJoke(c *gin.Context) {
	param := c.Query("joke_name")
	newUUID := uuid.New().String()
	ctx := context.WithValue(c.Request.Context(), "UUID", newUUID)

	fmt.Println("Finding joke with name: ", param)
	retrievedJoke, retrieveErr := GetJokeByName(ctx, param)
	if retrieveErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": retrieveErr.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": retrievedJoke,
	})
}

func GetAllJokes(c *gin.Context) {
	newUUID := uuid.New().String()
	ctx := context.WithValue(c.Request.Context(), "UUID", newUUID)

	retrievedJokes, retrieveErr := GetAllJokesFromMongo(ctx)
	if retrieveErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": retrieveErr.Error()})
		return
	}

	c.JSON(200, gin.H{
		"all_jokes": retrievedJokes,
	})
}

func CreateJoke(c *gin.Context) {
	var jokeBody MongoJoke
	if err := c.Bind(&jokeBody); err != nil {
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

func GetRandom(c *gin.Context) {
	retrievedJoke, retrieveErr := GetRandomJoke()
	if retrieveErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": retrieveErr.Error()})
		return
	}

	c.JSON(200, gin.H{
		"awesome_joke": retrievedJoke,
	})
}

type RandomJokeResp struct {
	HttpError bool        `json:"error"`
	Category  string      `json:"category"`
	Type      string      `json:"type"`
	Joke      string      `json:"joke"`
	Flags     interface{} `json:"flags"`
	Id        int         `json:"id"`
	Safe      bool        `json:"safe"`
	Lang      string      `json:"lang"`
}

func GetRandomJoke() (joke string, err error) {
	resp, err := http.Get("https://sv443.net/jokeapi/v2/joke/Programming,Miscellaneous,Dark?format=json&blacklistFlags=nsfw,sexist,racist&type=single")
	if err != nil {
		fmt.Println("error calling random joke generator: ", err)
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	fmt.Println("joke received io resp body: ", string(body))

	jokeResp := RandomJokeResp{}
	unmarshallErr := json.Unmarshal(body, &jokeResp)
	if unmarshallErr != nil {
		return "", unmarshallErr
	}

	fmt.Println("retrieved joke: ", jokeResp.Joke)
	return jokeResp.Joke, nil
}

const mongoUri = "mongodb://root:example@mongo:27017/"

func ConnectToMongo(ctx context.Context) (collection *mongo.Collection, err error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		mongoUri,
	))

	//err = client.Ping(ctx, nil)

	if err != nil {
		fmt.Println("Error connecting to mongo: ", err)
		return nil, err
	}

	var dbName = "dadJokes"
	var collectionName = "jokes"
	collection = client.Database(dbName).Collection(collectionName)
	return collection, nil
}

func InsertJokeToDatabase(ctx context.Context, joke MongoJoke) (err error) {
	collection, connectErr := ConnectToMongo(ctx)
	if connectErr != nil {
		fmt.Println("Error connecting to Mongo", connectErr)
		return connectErr
	}
	fmt.Println("creating joke with name: ", joke.JokeName)
	joke.ID = primitive.NewObjectID()

	insertManyResult, err := collection.InsertOne(context.TODO(), joke)
	if err != nil {
		fmt.Println("Something went wrong trying to insert the new documents: ", err)
		return err
	}

	fmt.Println("documents successfully inserted", insertManyResult)
	fmt.Println("insert ID", insertManyResult.InsertedID)

	return nil
}

func GetJokeByName(ctx context.Context, keyword string) (retrievedJoke MongoJoke, err error) {
	collection, connectErr := ConnectToMongo(ctx)
	if connectErr != nil {
		fmt.Println("Error connecting to Mongo", connectErr)
		return MongoJoke{}, connectErr
	}

	//possible GetJokeByID endpoint code but not sure if it's necessary
	// parmID, err := primitive.ObjectIDFromHex(keyword)
	// if err != nil {
	// return CreateJokeBody{}, err
	// }

	var result MongoJoke
	myFilter := bson.M{"joke_name": keyword}
	findErr := collection.FindOne(ctx, myFilter).Decode(&result)
	if findErr != nil {
		fmt.Println("Something went wrong trying to find one document: ", findErr)
		return MongoJoke{}, findErr
	}

	fmt.Println("Found a document", result)
	jokeBody := MongoJoke{
		ID:          result.ID,
		JokeName:    result.JokeName,
		JokeContent: result.JokeContent,
	}
	return jokeBody, nil
}

func GetAllJokesFromMongo(ctx context.Context) (allJokes []MongoJoke, err error) {
	collection, connectErr := ConnectToMongo(ctx)
	if connectErr != nil {
		fmt.Println("Error connecting to Mongo", connectErr)
		return []MongoJoke{}, connectErr
	}

	cursor, findErr := collection.Find(ctx, bson.M{})
	if findErr != nil {
		fmt.Println("Something went wrong trying to find one document: ", findErr)
		return []MongoJoke{}, findErr
	}
	defer cursor.Close(ctx)

	arrJokes := []MongoJoke{}
	for cursor.Next(ctx) {
		var result MongoJoke
		if err = cursor.Decode(&result); err != nil {
			fmt.Println("Error decoding result: ", err)
		}
		arrJokes = append(arrJokes, result)
	}
	return arrJokes, nil
}
