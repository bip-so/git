package models

import "time"

type Log struct {
	ID          string    `json:"id"`
	Message     string    `json:"message"`
	AuthorEmail string    `json:"authorEmail"`
	CreatedAt   time.Time `json:"createdAt"`
}
