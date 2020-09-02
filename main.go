package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var client *mongo.Client

type Meeting struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	StartTime    string `json:"starttime"`
	EndTime      string `json:"endtime"`
	CreationTime string `json:"creationtime"`
}

func InsertNewMeeting(meeting Meeting) interface{} {
	collection := client.Database("datab").Collection("meetings")
	insertResult, err := collection.InsertOne(context.TODO(), meeting)
	if err != nil {
		log.Fatalln("Error on inserting new Meeting", err)
	}
	return insertResult.InsertedID
}

func ReturnAllMeetings(filter bson.M) []*Meeting {
	var meetings []*Meeting
	collection := client.Database("datab").Collection("meetings")
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		log.Fatal("Error on Finding all the documents", err)
	}
	for cur.Next(context.TODO()) {
		var meeting Meeting
		err = cur.Decode(&meeting)
		if err != nil {
			log.Fatal("Error on Decoding the document", err)
		}
		meetings = append(meetings, &meeting)
	}
	return meetings
}
func GetClient() *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func InsertNewMeetingHandler(w http.ResponseWriter, r *http.Request) {
	// response.Header().Set("content-type", "application/json")
	// var meeting Meeting
	// _ = json.NewDecoder(request.Body).Decode(&meeting)
	// collection := client.Database("datab").Collection("meetings")
	// ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	// result, _ := collection.InsertOne(ctx, meeting)
	// json.NewEncoder(response).Encode(result)
	query := r.URL.Query()
	start := "00000"
	start = query.Get("start") //filters=["color", "price", "brand"]
	end := query.Get("end")
	if start == "00000" {
		fmt.Println("haaan")
		bodyBytes, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		ct := r.Header.Get("content-type")
		if ct != "application/json" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			w.Write([]byte(fmt.Sprintf("need content-type 'application/json', but got '%s'", ct)))
			return
		}

		var meeting Meeting
		err = json.Unmarshal(bodyBytes, &meeting)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		meeting.ID = fmt.Sprintf("%d", time.Now().UnixNano())
		meeting.CreationTime = time.Now().Format("2006-01-02 15:04:05")
		InsertNewMeeting(meeting)
	} else {

		meetings := ReturnAllMeetings(bson.M{"starttime": start, "endtime": end})
		jsonByte, err := json.Marshal(meetings)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonByte)
	}

}

func ReturnAllMeetingsHandler(w http.ResponseWriter, r *http.Request) {
	// response.Header().Set("content-type", "application/json")
	// params := mux.Vars(request)
	// id, _ := primitive.ObjectIDFromHex(params["id"])
	// var person Person
	// collection := client.Database("thepolyglotdeveloper").Collection("people")
	// ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	// err := collection.FindOne(ctx, Person{ID: id}).Decode(&person)
	// if err != nil {
	// 	response.WriteHeader(http.StatusInternalServerError)
	// 	response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
	// 	return
	// }
	// json.NewEncoder(response).Encode(person)
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fmt.Println(parts[2])
	meetings := ReturnAllMeetings(bson.M{"title": parts[2]})

	jsonByte, err := json.Marshal(meetings)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonByte)
}

func SearchMeeting(w http.ResponseWriter, r *http.Request) {
	// response.Header().Set("content-type", "application/json")
	// params := mux.Vars(request)
	// id, _ := primitive.ObjectIDFromHex(params["id"])
	// var person Person
	// collection := client.Database("thepolyglotdeveloper").Collection("people")
	// ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	// err := collection.FindOne(ctx, Person{ID: id}).Decode(&person)
	// if err != nil {
	// 	response.WriteHeader(http.StatusInternalServerError)
	// 	response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
	// 	return
	// }
	// json.NewEncoder(response).Encode(person)

	// meetings := ReturnAllMeetings(bson.M{"title": parts[2]})

	// jsonByte, err := json.Marshal(meetings)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	w.Write([]byte(err.Error()))
	// }

	// w.Header().Add("content-type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// w.Write(jsonByte)
}

func main() {
	c := GetClient()
	client = c
	err := c.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal("Couldn't connect to the database", err)
	} else {
		log.Println("Connected!")
	}

	http.HandleFunc("/meeting", InsertNewMeetingHandler)
	http.HandleFunc("/meeting/", ReturnAllMeetingsHandler)

	err = http.ListenAndServe(":8090", nil)
	if err != nil {
		panic(err)
	}
}
