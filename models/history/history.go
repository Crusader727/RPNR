package history

import (
	"time"

	"../../db"
	"github.com/jinzhu/gorm"
)

type Notation struct {
	gorm.Model

	Number string
	Img    string
	ImgRes float64
	UserID uint `gorm:"not null;index"`
}

func (n *Notation) Save() error {
	return db.Get().Create(n).Error
}

func (n *Notation) Get(user uint) error {
	return db.Get().Where("user_id = ?", user).First(n).Error
}

func GetLatest(user, limit uint, notations *[]Notation) interface{} {
	return db.Get().Limit(limit).Where("user_id = ?", user).Order("created_at desc").Find(notations).Error
}

func (n *Notation) GetFormatedCreatedAt() string {
	return n.CreatedAt.Format(time.Stamp)
}

func init() {
	db.Get().AutoMigrate(&Notation{})
}
