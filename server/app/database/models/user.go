package models

import (
	"github.com/google/uuid"
	"go-scrape-this/server/app/database/structs"
	"gorm.io/gorm"
	"time"
)

// calibrate with: "docker run -it --rm --entrypoint kratos oryd/kratos:v0.5 hashers argon2 calibrate 1s"
var passwordParams = structs.NewPasswordParams(
	4194304,
	1,
	64,
	16,
	32,
)

type User struct {
	ID        uuid.UUID            `gorm:"primaryKey,type:string,size:36,<-:create" json:"id"`
	Username  string               `gorm:"unique" json:"username"`
	Password  structs.PasswordHash `gorm:"type:string" json:"-"`
	CreatedAt time.Time            `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt time.Time            `gorm:"autoUpdateTime:milli" json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt       `gorm:"index" json:"deleted_at,omitempty"`
}

func NewUser(username string, password string) (User, error) {
	passwordHash, err := structs.NewPasswordHash(password, passwordParams)
	if err != nil {
		return User{}, err
	}
	return User{
		ID:       uuid.New(),
		Username: username,
		Password: passwordHash,
	}, nil
}
