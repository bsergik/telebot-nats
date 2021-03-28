package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/******************************************************************************
 * Types
 */

type storage struct {
	database *gorm.DB
}

/******************************************************************************
 * Functions
 */

func Connect(dbHost, dbName, dbUser, dbPassword string, dbPort uint) (db storage, err error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable port=%d host=%s",
		dbUser, dbPassword, dbName, dbPort, dbHost)

	db.database, err = gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})

	return db, err
}

func (db storage) Disconnect() {
	if db.database != nil {
		sqlDB, _ := db.database.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
}

func (db storage) IsConnected() bool {
	return db.database != nil
}

func (db storage) InitDB() (err error) {
	tables := []interface{}{
		new(Message),
		new(Recipient),
	}

	for i := range tables {
		err = db.database.AutoMigrate(tables[i])

		if err != nil {
			return err
		}
	}

	return nil
}

func (db storage) AddMessage(msg *Message) (err error) {
	return db.database.Select("message", "created_at", "sent_at").Create(&msg).Error
}

func (db storage) AddRecipient(id int) (err error) {
	rcp := &Recipient{
		ID:          0,
		RecipientID: uint64(id),
		CreatedAt:   time.Now(),
	}

	return db.database.Create(rcp).Error
}

func (db storage) RemoveRecipient(id int) (err error) {
	return db.database.Delete(new(Recipient), "recipient_id = ?", id).Error
}

func (db storage) GetRecipients() (rcpns []Recipient, err error) {
	rcpns = make([]Recipient, 0)

	return rcpns, db.database.Model(new(Recipient)).Find(&rcpns).Error
}
