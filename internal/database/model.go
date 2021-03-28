package database

import (
	"time"
)

type Message struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement;type:bigserial"`
	Subsystem string    `gorm:"column:subsystem;not null;type:character varying[100]"`
	Message   string    `gorm:"column:message;not null;type:character varying"`
	CreatedAt time.Time `gorm:"column:created_at;not null;type:timestamp with time zone"`
	SentAt    time.Time `gorm:"column:sent_at;type:timestamp with time zone"`
}

type Recipient struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement;type:bigserial"`
	RecipientID uint64    `gorm:"column:recipient_id;unique;not null;type:bigint"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;type:timestamp with time zone"`
}
