package user

import (
	"../../db"
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model

	Username string `gorm:"not null;unique_index"`
	Password string
}

func (u *User) GetUserByName() error {
	return db.Get().Where("username = ?", u.Username).First(u).Error
}

func (u *User) Create() error {
	return db.Get().Create(u).Error
}

func init() {
	db.Get().AutoMigrate(&User{})
}
