package models

import (
	"github.com/google/uuid"
	"go-scrape-this/server/app/database/structs"
	"gorm.io/gorm"
	"time"
)

var passwordParams = structs.NewPasswordParams(
	524288,
	3,
	4,
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
