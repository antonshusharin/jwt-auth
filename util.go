package main

import (
	"log"
	"os"
	"testing"
)

func GetEnvOptional(name, defaultVal string) string {
	if testing.Testing() {
		name += "_TESTING"
	}
	val, ok := os.LookupEnv(name)
	if !ok || val == "" {
		return defaultVal
	}
	return val
}

func GetEnvRequired(name string) string {
	if testing.Testing() {
		name += "_TESTING"
	}
	val := os.Getenv(name)
	if val == "" {
		log.Fatalf("Error: the environment variable '%v' must be set", name)
	}
	return val
}
