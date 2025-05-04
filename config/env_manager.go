package config

import "os"

type Env struct {
	MongoURI        string
	MongoDBLearnWiz string
	ServerPort      string
}

var EnvManager *Env

func NewEnvManager() *Env {
	return &Env{
		MongoURI:        os.Getenv("MONGO_URI"),
		MongoDBLearnWiz: os.Getenv("MONGO_DB_LEARN_WIZ"),
		ServerPort:      os.Getenv("SERVER_PORT"),
	}
}
