package model

import "time"

type UserRecord struct {
	ID           int64     `db:"id"`
	Email        string    `db:"type:varchar(255);unique_index"`
	PasswordHash string    `db:"type:varchar(255)"`
	CreatedAt    time.Time `db:"autoCreateTime"`
}
