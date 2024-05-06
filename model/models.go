package model

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID           primitive.ObjectID    `bson:"_id"`
	Name         string    `bson:"name"`
	Username     string    `bson:"username" unique:"true"`
	ProfilePicUrl string   `bson:"profilePicUrl,omitempty"`
	Password     string    `bson:"password"`
	CreatedAt    time.Time `bson:"createdAt"`
	UpdatedAt    *time.Time `bson:"updatedAt,omitempty"`
}

type Channel struct {
	ID          primitive.ObjectID    `bson:"_id"`
	Name        string    `bson:"name,omitempty"`
	Description string    `bson:"description,omitempty"`
	IconUrl     string    `bson:"iconUrl,omitempty"`
	Type        string    `bson:"type" default:"DIRECT_MSG"`
	LastMsg     string    `bson:"lastMsg,omitempty"`
	CreatedAt   time.Time `bson:"createdAt"`
	UpdatedAt   *time.Time `bson:"updatedAt,omitempty"`
	Members     []primitive.ObjectID `bson:"members"`
}

type Message struct {
	ID        primitive.ObjectID    `bson:"_id"`
	ChannelID primitive.ObjectID    `bson:"channelId"`
	CreatedBy primitive.ObjectID    `bson:"createdBy"`
	Content   string    `bson:"content"`
	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt *time.Time `bson:"updatedAt,omitempty"`
}
