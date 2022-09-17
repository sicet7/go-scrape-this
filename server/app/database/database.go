package database

import (
	"database/sql"
	"errors"
	"github.com/rs/zerolog"
	"go-scrape-this/server/app/database/models"
	"golang.org/x/exp/maps"
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

func (s DatabaseType) String() string {
	return s.value
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
	databaseModels = map[string]interface{}{
		"user": models.User{},
	}
)

type Database struct {
	conn *gorm.DB
}

func NewDatabase(dbType *DatabaseType, dsn string, providedLogger *zerolog.Logger) (Database, error) {

	if !slices.Contains(supported, dbType) {
		return Database{}, errors.New("unknown or unsupported database type")
	}

	gormLogging := gormLogger.New(providedLogger, gormLogger.Config{})

	var connection *gorm.DB
	switch dbType {
	case SQLITE:
		dbCon, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: gormLogging,
		})
		if err != nil {
			return Database{}, err
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
			return Database{}, err
		}
		connection = dbCon
		break
	case POSTGRESQL:
		dbCon, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogging,
		})
		if err != nil {
			return Database{}, err
		}
		connection = dbCon
		break
	case MYSQL:
		sqlDB, err := sql.Open("mysql", dsn)
		if err != nil {
			return Database{}, err
		}
		dbCon, err := gorm.Open(mysql.New(mysql.Config{
			Conn:              sqlDB,
			DefaultStringSize: 256,
		}), &gorm.Config{
			Logger: gormLogging,
		})
		if err != nil {
			return Database{}, err
		}
		connection = dbCon
		break
	default:
		return Database{}, errors.New("unknown or unsupported database type")
	}

	db := Database{
		conn: connection,
	}

	return db, nil
}

func (d *Database) GetConnection() *gorm.DB {
	return d.conn
}

func (d *Database) RunMigrations() error {
	return d.conn.AutoMigrate(maps.Values(databaseModels)...)
}

func (d *Database) Model(name string) {
	//d.GetConnection().Model()
}
