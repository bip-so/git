package models

import "time"

type Attribution struct {
	AuthorEmail string `json:"authorEmail"`
	Edits       int    `json:"edits"`
}

type BlockAttribution struct {
	AuthorEmail string    `json:"authorEmail"`
	BlockID     string    `json:"blockId"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
