package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "gopoc/model"
    "gopoc/db"
    "log"
	"time"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type ChannelCreateInput struct {
    Member string `json:"member"`
}

func CreateChannel(w http.ResponseWriter, r *http.Request) {
    var dto ChannelCreateInput
    var channel model.Channel

    authorization := r.Header.Get("Authorization")[len("Bearer "):]

    err := json.NewDecoder(r.Body).Decode(&dto)
    if err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    memberID, err := primitive.ObjectIDFromHex(dto.Member)
    if err != nil {
        http.Error(w, "Invalid member ID format", http.StatusBadRequest)
        return
    }


	claims, err := ParseToken(authorization)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

    userID,err := primitive.ObjectIDFromHex(claims.UserID)
    if err != nil {
        http.Error(w, "Invalid user ID format", http.StatusBadRequest)
        return
    }
    
    collection := db.Client.Database("chatDB").Collection("channels")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err = collection.FindOne(context.TODO(), bson.M{
        "members": bson.M{
            "$all": []primitive.ObjectID{memberID, userID},
            "$size": 2,
        },
    }).Decode(&channel)

    if err != mongo.ErrNoDocuments {
        if err != nil {
            log.Printf("Failed to find channel: %v", err)
            http.Error(w, "Failed to find channel", http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(channel)
        return
    }
    
    channel.ID = primitive.NewObjectID()
    channel.Members = []primitive.ObjectID{memberID, userID}
    channel.Type = "DIRECT_MESSAGE"
    channel.CreatedAt = time.Now()

    _, err = collection.InsertOne(ctx, channel)
    if err != nil {
        log.Printf("Failed to create channel: %v", err)
        http.Error(w, "Failed to create channel", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(channel)
}
