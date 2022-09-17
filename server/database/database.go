package database

import (
	"database/sql"
	"errors"
	"github.com/rs/zerolog"
	"golang.org/x/exp/slices"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"strings"
)

type DatabaseType struct {
	value string
}

var (
	SQLITE     = &DatabaseType{value: "sqlite"}
	MSSQL      = &DatabaseType{value: "mssql"}
	POSTGRESQL = &DatabaseType{value: "pssql"}
	MYSQL      = &DatabaseType{value: "mysql"}
	supported  = []*DatabaseType{
		SQLITE,
		MSSQL,
		POSTGRESQL,
		MYSQL,
	}
)

func (s DatabaseType) String() string {
	return s.value
}

type Database struct {
	conn *gorm.DB
}

func (d *Database) GetConnection() *gorm.DB {
	return d.conn
}

func (d *Database) RunMigrations() {

}

func ParseDatabaseType(value string) (*DatabaseType, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case SQLITE.String():
		return SQLITE, nil
	case MSSQL.String():
		return MSSQL, nil
	case POSTGRESQL.String():
		return POSTGRESQL, nil
	case MYSQL.String():
		return MYSQL, nil
	}
	return nil, errors.New("unknown or unsupported database type")
}

func NewDatabase(dbType *DatabaseType, dsn string, providedLogger *zerolog.Logger) (*Database, error) {

	if !slices.Contains(supported, dbType) {
		return nil, errors.New("unknown or unsupported database type")
	}

	gormLogging := gormLogger.New(providedLogger, gormLogger.Config{})

	var connection *gorm.DB
	switch dbType {
	case SQLITE:
		dbCon, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: gormLogging,
		})
		if err != nil {
			return nil, err
		}
		connection = dbCon
		break
	case MSSQL:
		dbCon, err := gorm.Open(sqlserver.New(sqlserver.Config{
			DSN:               dsn,
			DefaultStringSize: 256,
		}), &gorm.Config{
			Logger: gormLogging,
		})
		if err != nil {
			return nil, err
		}
		connection = dbCon
		break
	case POSTGRESQL:
		dbCon, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogging,
		})
		if err != nil {
			return nil, err
		}
		connection = dbCon
		break
	case MYSQL:
		sqlDB, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, err
		}
		dbCon, err := gorm.Open(mysql.New(mysql.Config{
			Conn:              sqlDB,
			DefaultStringSize: 256,
		}), &gorm.Config{
			Logger: gormLogging,
		})
		if err != nil {
			return nil, err
		}
		connection = dbCon
		break
	default:
		return nil, errors.New("unknown or unsupported database type")
	}

	db := &Database{
		conn: connection,
	}

	db.RunMigrations()

	return db, nil
}
