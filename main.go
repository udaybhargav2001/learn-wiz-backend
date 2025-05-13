package main

import (
	"fmt"
	"learn-wiz-backend/config"
	"log"
	"net"

	quizpb "learn-wiz-backend/internal/api/pb"
	"learn-wiz-backend/internal/api/quiz"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Couldnt load env file")
	}
	config.EnvManager = config.NewEnvManager()
}
func main() {
	fmt.Println("Hello, World!")
	config.InitMongoDB()

	//init grpc server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.EnvManager.ServerPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	quizServer := quiz.NewServer()
	quizpb.RegisterQuizServiceServer(grpcServer, quizServer)

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
