package database

import (
	"database/sql"
	"golang.org/x/exp/slices"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"log"
	"os"
	"strings"
)

var types []string = []string{
	"sqlite",
	"mssql",
	"mysql",
	"pssql",
}
var conn *gorm.DB = nil

func Get() *gorm.DB {
	if conn == nil {
		log.Fatal("Tried to access uninitialized database connection")
	}
	return conn
}

func Init() {
	if conn != nil {
		log.Fatal("Tried to re-initialized database connection")
	}

	dbType, ok := os.LookupEnv("DATABASE_TYPE")
	if !ok {
		dbType = "sqlite"
	}

	if !slices.Contains(types, dbType) {
		log.Fatalf("Unknown or unsupported database type, supported types include: \"%s\"", strings.Trim(strings.Join(types, "\",\""), "\","))
	}

	dbDsn, ok := os.LookupEnv("DATABASE_DSN")
	if !ok {
		if dbType != "sqlite" {
			log.Fatal("Missing environment variable \"DATABASE_DSN\" to connect to database.")
		}
		dbDsn = "scraper.db"
	}

	switch dbType {
	case "sqlite":
		db, err := gorm.Open(sqlite.Open(dbDsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to Sqlite database: %s", err.Error())
		}
		conn = db
	case "mssql":
		db, err := gorm.Open(sqlserver.New(sqlserver.Config{
			DSN:               dbDsn,
			DefaultStringSize: 256,
		}), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to Microsoft SQL database: %s", err.Error())
		}
		conn = db
	case "pssql":
		db, err := gorm.Open(postgres.Open(dbDsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL database: %s", err.Error())
		}
		conn = db
	case "mysql":
		sqlDB, err := sql.Open("mysql", dbDsn)
		if err != nil {
			log.Fatalf("Failed to create MySQL connector: %s", err.Error())
		}
		db, err := gorm.Open(mysql.New(mysql.Config{
			Conn:              sqlDB,
			DefaultStringSize: 256,
		}), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to MySQL server: %s", err.Error())
		}
		conn = db
	}
}
