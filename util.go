package main

import (
	"log"
	"os"
)

func GetEnvOptional(name, defaultVal string) string {
	val, ok := os.LookupEnv(name)
	if !ok || val == "" {
		return defaultVal
	}
	return val
}

func GetEnvRequired(name string) string {
	val := os.Getenv(name)
	if val == "" {
		log.Fatalf("Error: the environment variable '%v' must be set", name)
	}
	return val
}
