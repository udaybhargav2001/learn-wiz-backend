package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Question represents one MCQ in the question bank.
// DifficultyRating follows Elo-like rating; default 1000.
// CorrectIndex is stored server-side, not exposed in gRPC Question message.

type Question struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	Stem             string             `bson:"stem"`
	Options          []string           `bson:"options"`
	CorrectIndex     int32              `bson:"correct_index"`
	DifficultyRating float64            `bson:"difficulty_rating"`
	TopicID          primitive.ObjectID `bson:"topic_id"`
}
