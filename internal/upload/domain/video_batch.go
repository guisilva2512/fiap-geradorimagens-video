package domain

import "time"

type VideoBatch struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateVideoBatchCommand colocamos a struct de comando aqui dentro do domínio.
type CreateVideoBatchCommand struct {
	UserID string
}
