package db

import "os"

var Pool interface{}

func Init() error {
	_ = os.Getenv("DATABASE_URL")
	return nil
}

func Close()             {}
func GetDB() interface{} { return Pool }
