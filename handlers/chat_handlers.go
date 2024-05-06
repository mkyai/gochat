package handlers

import (
	"context"
	"net/http"
	"gopoc/model"
	"gopoc/db"
	"encoding/json"
	"time"
	"log"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetChannels(w http.ResponseWriter, r *http.Request) {
	var results []bson.M

	authorization := r.Header.Get("Authorization")[len("Bearer "):]

	claims, err := ParseToken(authorization)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	collection := db.Client.Database("chatDB").Collection("channels")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	pipeline := mongo.Pipeline{
        {{"$match", bson.M{"members": userID}}},
        {{"$lookup", bson.M{
            "from": "users",
            "localField": "members",
            "foreignField": "_id",
			"pipeline": []bson.M{            
                {"$project": bson.M{ 
                    "_id": 1,
                    "name": 1,
                    "username": 1,
                }},
            },
            "as": "memberDetails",
        }}},

        {{"$project", bson.M{
            "name": 1,
			"lastMessage": 1,
			"createdAt": 1,
			"updatedAt": 1,
            "description": 1,
            "memberDetails": bson.M{"$filter": bson.M{
                "input": "$memberDetails",
                "as": "member",
                "cond": bson.M{"$ne": []interface{}{"$$member._id", userID}},
            }},
        }}},
    }

    cursor, err := collection.Aggregate(ctx, pipeline)
    if err != nil {
		log.Printf("Failed to fetch channels: %v", err)
        return 
    }
    defer cursor.Close(ctx)

    for cursor.Next(ctx) {
        var result bson.M
        if err := cursor.Decode(&result); err != nil {
			log.Printf("Failed to parse channels: %v", err)
            return 
        }
        results = append(results, result)
    }

    if err := cursor.Err(); err != nil {
		log.Printf("Failed to parse channels: %v", err)
        return 
    }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func WriteMessageToDB(cId primitive.ObjectID, uId primitive.ObjectID,m string) error {
	var msg model.Message

	msg.ID = primitive.NewObjectID()
	msg.ChannelID = cId
	msg.CreatedBy = uId
	msg.Content = m
	msg.CreatedAt = time.Now()

	collection := db.Client.Database("chatDB").Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}

func GetMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, er := primitive.ObjectIDFromHex(vars["channelID"])
	if er != nil {
		log.Printf("Invalid channel ID format: %v", er)
		return
	}

	collection := db.Client.Database("chatDB").Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var messages []model.Message
	filter := bson.M{"channelId": channelID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &messages); err != nil {
		http.Error(w, "Failed to parse messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func SendMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, er := primitive.ObjectIDFromHex(vars["channelID"])
	if er != nil {
		log.Printf("Invalid channel ID format: %v", er)
		return
	}	

	var msg model.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	msg.ChannelID = channelID
	msg.CreatedAt = time.Now()

	collection := db.Client.Database("chatDB").Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, msg)
	if err != nil {
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}