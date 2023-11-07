package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email  string
	Name   string
	Audios []Audio
}

type Audio struct {
	gorm.Model
	Name           string
	Path           string
	TranscriptPath string
	UserID         uint
}

func (a *Audio) HasTranscript() bool {
	return a.TranscriptPath != ""
}

// http://go-database-sql.org/retrieving.html

func initDb() {

	db, err := gorm.Open(sqlite.Open("./database/db"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	// Migrate schemas
	err = db.AutoMigrate(&User{})
	if err != nil {
		panic(fmt.Sprintf("Failed to migrate User table: %s", err))
	}
	err = db.AutoMigrate(&Audio{})
	if err != nil {
		panic(fmt.Sprintf("Failed to migrate Audio table: %s", err))
	}
}

func getDb() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open("./database/db"), &gorm.Config{})
}
