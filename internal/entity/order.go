package entity

import "time"

type Order struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expiresAt"`
	Version   int64     `json:"version"`
	UserID    int64     `json:"userId"`
	*Ticket   `json:"ticket"`
}
