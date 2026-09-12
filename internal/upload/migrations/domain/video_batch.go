package domain

import "time"

type VideoBatch struct {
	ID        string
	IDUser    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateVideoBatchCommand colocamos a struct de comando aqui dentro do domínio.
type CreateVideoBatchCommand struct {
	IDUser string
}
