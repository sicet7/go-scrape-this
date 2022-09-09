package database

import (
	"database/sql"
	"errors"
	"github.com/rs/zerolog"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"os"
)

type Database struct {
	conn *gorm.DB
}

func (d *Database) GetConnection() *gorm.DB {
	return d.conn
}

func (d *Database) RunMigrations() {

}

func NewDatabase(providedLogger *zerolog.Logger) (*Database, error) {
	connection, err := createConnection(providedLogger)
	if err != nil {
		return nil, err
	}
	db := &Database{
		conn: connection,
	}

	db.RunMigrations()

	return db, nil
}

func createConnection(providedLogger *zerolog.Logger) (*gorm.DB, error) {
	dbType, ok := os.LookupEnv("DATABASE_TYPE")
	if !ok {
		dbType = "sqlite"
	}

	dbDsn, ok := os.LookupEnv("DATABASE_DSN")
	if !ok {
		if dbType != "sqlite" {
			return nil, errors.New("missing environment variable \"DATABASE_DSN\" to connect to database")
		}
		dbDsn = "scraper.db"
	}

	gormLogging := gormLogger.New(providedLogger, gormLogger.Config{})

	switch dbType {
	case "sqlite":
		db, err := gorm.Open(sqlite.Open(dbDsn), &gorm.Config{
			Logger: gormLogging,
		})
		if err != nil {
			return nil, err
		}
		return db, nil
	case "mssql":
		db, err := gorm.Open(sqlserver.New(sqlserver.Config{
			DSN:               dbDsn,
			DefaultStringSize: 256,
		}), &gorm.Config{
			Logger: gormLogging,
		})
		if err != nil {
			return nil, err
		}
		return db, nil
	case "pssql":
		db, err := gorm.Open(postgres.Open(dbDsn), &gorm.Config{
			Logger: gormLogging,
		})
		if err != nil {
			return nil, err
		}
		return db, nil
	case "mysql":
		sqlDB, err := sql.Open("mysql", dbDsn)
		if err != nil {
			return nil, err
		}
		db, err := gorm.Open(mysql.New(mysql.Config{
			Conn:              sqlDB,
			DefaultStringSize: 256,
		}), &gorm.Config{
			Logger: gormLogging,
		})
		if err != nil {
			return nil, err
		}
		return db, nil
	}
	return nil, errors.New("unknown or unsupported database type")
}
