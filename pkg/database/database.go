package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

// NewSQLiteDB creates a new gorm.DB from given db path.
func NewSQLiteDB(dbpath string) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)

	l := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	for i := 0; i < 10; i++ {
		db, err = gorm.Open(sqlite.Open(dbpath), &gorm.Config{
			Logger: l,
		})
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		return nil, err
	}
	return db, nil
}
