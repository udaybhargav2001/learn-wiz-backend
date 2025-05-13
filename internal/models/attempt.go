package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Attempt struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	StudentID          primitive.ObjectID `bson:"student_id"`
	QuestionID         primitive.ObjectID `bson:"question_id"`
	ChosenIndex        int32              `bson:"chosen_index"`
	Correct            bool               `bson:"correct"`
	TimeTakenMs        int64              `bson:"time_taken_ms"`
	AttemptedAt        time.Time          `bson:"attempted_at"`
	StudentRatingPre   float64            `bson:"student_rating_pre"`
	StudentRatingPost  float64            `bson:"student_rating_post"`
	QuestionRatingPre  float64            `bson:"question_rating_pre"`
	QuestionRatingPost float64            `bson:"question_rating_post"`
}
