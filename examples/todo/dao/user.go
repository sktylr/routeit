package dao

import "time"

type User struct {
	Id       string
	Created  time.Time
	Updated  time.Time
	Name     string
	Email    string
	Password string
}
