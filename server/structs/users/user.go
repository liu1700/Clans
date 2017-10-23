package users

import (
	"Clans/server/db"
	"Clans/server/log"
	"time"

	"github.com/jinzhu/gorm"
)

type User struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt int64
	UpdatedAt int64
	Name      string
	PassWord  string
}

func FindUserByName(name string) *User {
	u := new(User)
	if err := db.DB().Where(&User{Name: name}).Find(u).Error; err == nil {
		return u
	} else if err != gorm.ErrRecordNotFound {
		log.Logger().Error("error when FindUserByName err ", err.Error())
	}
	return nil
}

func CreateUser(name string, pw string) *User {
	u := User{Name: name}

	now := time.Now().Unix()
	newUser := User{
		Name:      name,
		PassWord:  pw,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.DB().Where(&u).FirstOrCreate(&u, newUser).Error; err != nil {
		log.Logger().Error("error when CreateUser err ", err.Error())
		return nil
	}
	return &u
}
