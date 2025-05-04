package main

import (
	"fmt"
	"learn-wiz-backend/config"
	"log"

	"github.com/joho/godotenv"
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

}
