package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Article Structure
type Article struct {
	ID       string    `json:"id,omitempty" bson:"_id,omitempty"`
	Title    string    `json:"title,omitempty" bson:"title,omitempty"`
	SubTitle string    `json:"subtitle,omitempty" bson:"subtitle,omitempty"`
	Content  string    `json:"content,omitempty" bson:"content,omitempty"`
	Created  time.Time `json:"created,omitempty" bson:"created,omitempty"`
}

var client *mongo.Client

// Connecting with the database (MongoDB)
func connect() {

	clientOptions := options.Client().ApplyURI("mongodb://himsngh:H22imanshu04@goapicluster-shard-00-00.2usbp.mongodb.net:27017,goapicluster-shard-00-01.2usbp.mongodb.net:27017,goapicluster-shard-00-02.2usbp.mongodb.net:27017/test?ssl=true&replicaSet=atlas-b5saj8-shard-0&authSource=admin&retryWrites=true&w=majority")
	client, _ = mongo.NewClient(clientOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.Background(), readpref.Primary())

	if err != nil {
		log.Fatal("Couldn't connect to the database", err)
	} else {
		log.Println("Connected to MondoDB Server")
	}

}

// Home Page
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello Himanshu!")
	fmt.Println("Endopoint Hit: Home Page")
}

// Function for returning all articles and posting articles
func returnAllArticles(response http.ResponseWriter, request *http.Request) {

	if request.Method == "GET" {
		var articles []Article
		collection := client.Database("test").Collection("Article")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var article Article
			cursor.Decode(&article)
			articles = append(articles, article)
		}
		if err = cursor.Err(); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		fmt.Println("Endpoint Hit: returnAllArticles")
		json.NewEncoder(response).Encode(articles)
	} else {
		request.ParseForm()
		decoder := json.NewDecoder(request.Body)
		var newArticle Article
		newArticle.Created = time.Now()
		err := decoder.Decode(&newArticle)
		if err != nil {
			panic(err)
		}
		log.Println(newArticle.ID)
		fmt.Println("Endpoint Hit: Post Article")
		insertArticle(newArticle)
	}
}

// For querying articles on id
func returnSingleArticle(response http.ResponseWriter, request *http.Request) {

	request.ParseForm()
	var id string = request.URL.Path
	id = id[10:]
	var article Article
	collection := client.Database("test").Collection("Article")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, Article{ID: id}).Decode(&article)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	fmt.Println("Returned Article ID NO : ", article.ID)
	json.NewEncoder(response).Encode(article)
}

// For query the database using the search query q=
func returnSearchResult(response http.ResponseWriter, request *http.Request) {
	fmt.Println("Searching.....")
	q := request.FormValue("q")
	collection := client.Database("test").Collection("Article")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// creating a new index for search query
	index := mongo.IndexModel{Keys: bson.M{"title": "text", "content": "text", "subtitle": "text"}}
	if _, err := collection.Indexes().CreateOne(ctx, index); err != nil {
		log.Println("Could not create index:", err)
	}
	cursor, err := collection.Find(ctx, bson.M{"$text": bson.M{"$search": q}})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	var articles []Article
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var article Article
		cursor.Decode(&article)
		articles = append(articles, article)
	}
	if err = cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	fmt.Println("Endpoint Hit: returnAllArticles")
	json.NewEncoder(response).Encode(articles)
}

// function for inserting the article in the database
func insertArticle(article Article) {
	collection := client.Database("test").Collection("Article")
	insertResult, err := collection.InsertOne(context.TODO(), article)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted post with ID:", insertResult.InsertedID)
}

// function for handling the request from the client.
func handleRequest() {

	http.HandleFunc("/", homePage)
	http.HandleFunc("/articles", returnAllArticles)
	http.HandleFunc("/articles/", returnSingleArticle)
	http.HandleFunc("/articles/search", returnSearchResult)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe", err)
	}
}

func main() {
	connect()
	handleRequest()
}
