package quiz

import (
	"context"
	"errors"
	"math"
	"time"

	"learn-wiz-backend/config"
	pb "learn-wiz-backend/internal/api/pb"
	"learn-wiz-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	defaultRating = 1000.0
	kStudent      = 32.0
	kQuestion     = 32.0
	ratingWindow  = 75.0
)

// Server implements quizpb.QuizServiceServer
// Mongo collections:
//  - questions
//  - students  (stores rating)
//  - attempts

type Server struct {
	pb.UnimplementedQuizServiceServer
	questionsCol *mongo.Collection
	studentsCol  *mongo.Collection
	attemptsCol  *mongo.Collection
}

func NewServer() *Server {
	db := config.GetLwDB()
	return &Server{
		questionsCol: db.Collection("questions"),
		studentsCol:  db.Collection("students"),
		attemptsCol:  db.Collection("attempts"),
	}
}

// helper to fetch student rating (create if missing)
func (s *Server) getStudentRating(ctx context.Context, studentID primitive.ObjectID) (float64, error) {
	var doc struct {
		ID     primitive.ObjectID `bson:"_id"`
		Rating float64            `bson:"rating"`
	}
	err := s.studentsCol.FindOne(ctx, bson.M{"_id": studentID}).Decode(&doc)
	if err != nil {
		// if not found, create with default rating
		if errors.Is(err, mongo.ErrNoDocuments) {
			_, err2 := s.studentsCol.InsertOne(ctx, bson.M{"_id": studentID, "rating": defaultRating})
			if err2 != nil {
				return 0, err2
			}
			return defaultRating, nil
		}
		return 0, err
	}
	return doc.Rating, nil
}

func (s *Server) updateStudentRating(ctx context.Context, studentID primitive.ObjectID, newRating float64) error {
	_, err := s.studentsCol.UpdateByID(ctx, studentID, bson.M{"$set": bson.M{"rating": newRating}})
	return err
}

func (s *Server) getQuestionByDifficulty(ctx context.Context, topicID primitive.ObjectID, min, max float64) (*models.Question, error) {
	cursor, err := s.questionsCol.Aggregate(ctx, bson.A{
		bson.M{"$match": bson.M{"topic_id": topicID, "difficulty_rating": bson.M{"$gte": min, "$lte": max}}},
		bson.M{"$sample": bson.M{"size": 1}},
	})
	if err != nil {
		return nil, err
	}
	var questions []models.Question
	if err := cursor.All(ctx, &questions); err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return &questions[0], nil
}

func modelToProto(q *models.Question) *pb.Question {
	return &pb.Question{
		Id:               q.ID.Hex(),
		Stem:             q.Stem,
		Options:          q.Options,
		DifficultyRating: q.DifficultyRating,
		TopicId:          q.TopicID.Hex(),
	}
}

func (s *Server) GetNextQuestion(ctx context.Context, req *pb.GetNextQuestionRequest) (*pb.QuestionResponse, error) {
	studentObjID, err := primitive.ObjectIDFromHex(req.GetStudentId())
	if err != nil {
		return nil, err
	}
	topicObjID, err := primitive.ObjectIDFromHex(req.GetTopicId())
	if err != nil {
		return nil, err
	}

	studentRating, err := s.getStudentRating(ctx, studentObjID)
	if err != nil {
		return nil, err
	}

	min := studentRating - ratingWindow
	max := studentRating + ratingWindow

	question, err := s.getQuestionByDifficulty(ctx, topicObjID, min, max)
	if err != nil {
		// fallback: easiest question in topic
		if errors.Is(err, mongo.ErrNoDocuments) {
			var q models.Question
			err = s.questionsCol.FindOne(ctx, bson.M{"topic_id": topicObjID}).Decode(&q)
			if err != nil {
				return nil, err
			}
			question = &q
		} else {
			return nil, err
		}
	}

	return &pb.QuestionResponse{Question: modelToProto(question)}, nil
}

func expectedScore(rA, rB float64) float64 {
	return 1.0 / (1.0 + math.Pow(10, (rB-rA)/400))
}

func newRating(oldRating float64, k float64, score, expected float64) float64 {
	return oldRating + k*(score-expected)
}

func (s *Server) SubmitAnswer(ctx context.Context, req *pb.SubmitAnswerRequest) (*pb.SubmitAnswerResponse, error) {
	studentObjID, err := primitive.ObjectIDFromHex(req.GetStudentId())
	if err != nil {
		return nil, err
	}
	questionObjID, err := primitive.ObjectIDFromHex(req.GetQuestionId())
	if err != nil {
		return nil, err
	}

	// fetch question
	var question models.Question
	if err := s.questionsCol.FindOne(ctx, bson.M{"_id": questionObjID}).Decode(&question); err != nil {
		return nil, err
	}

	studentRating, err := s.getStudentRating(ctx, studentObjID)
	if err != nil {
		return nil, err
	}

	// Determine correctness
	correct := req.GetChosenIndex() == question.CorrectIndex

	// Elo calculations
	expectedStudent := expectedScore(studentRating, question.DifficultyRating)
	scoreStudent := 0.0
	scoreQuestion := 1.0
	if correct {
		scoreStudent = 1.0
		scoreQuestion = 0.0
	}

	newStudentRating := newRating(studentRating, kStudent, scoreStudent, expectedStudent)
	newQuestionRating := newRating(question.DifficultyRating, kQuestion, scoreQuestion, 1-expectedStudent)

	// Update DB
	_ = s.updateStudentRating(ctx, studentObjID, newStudentRating)
	_, _ = s.questionsCol.UpdateByID(ctx, question.ID, bson.M{"$set": bson.M{"difficulty_rating": newQuestionRating}})

	// Insert attempt
	attempt := models.Attempt{
		StudentID:          studentObjID,
		QuestionID:         questionObjID,
		ChosenIndex:        req.GetChosenIndex(),
		Correct:            correct,
		TimeTakenMs:        req.GetTimeTakenMs(),
		AttemptedAt:        time.Now(),
		StudentRatingPre:   studentRating,
		StudentRatingPost:  newStudentRating,
		QuestionRatingPre:  question.DifficultyRating,
		QuestionRatingPost: newQuestionRating,
	}
	_, _ = s.attemptsCol.InsertOne(ctx, attempt)

	// Fetch next question
	nextQResp, err := s.GetNextQuestion(ctx, &pb.GetNextQuestionRequest{
		StudentId: req.GetStudentId(),
		TopicId:   question.TopicID.Hex(),
	})
	if err != nil {
		// If error, just proceed without next question
		return &pb.SubmitAnswerResponse{
			Correct:          correct,
			NewStudentRating: newStudentRating,
			QuestionRating:   newQuestionRating,
		}, nil
	}

	return &pb.SubmitAnswerResponse{
		Correct:          correct,
		NewStudentRating: newStudentRating,
		QuestionRating:   newQuestionRating,
		NextQuestion:     nextQResp.Question,
	}, nil
}
