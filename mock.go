package main

import "github.com/google/uuid"

type MockUser struct {
	guid            uuid.UUID
	username, email string
}

var MOCK_USERS map[string]MockUser = map[string]MockUser{
	"Ivan":  {uuid.MustParse("5d174856-ec35-46d8-9e94-172ea2d30d04"), "Ivan", "ivan.ivanov@example.com"},
	"Maria": {uuid.MustParse("ca63f360-f41b-43a2-a1cc-32ecbf1c2aa3"), "Maria", "maria215314@example.com"},
}
