package thirdParty

import (
	"Clans/server/db"
	"Clans/server/log"

	"github.com/jinzhu/gorm"
)

const (
	PlatformAndroid = 1
	PlatformIOS     = 2
	PlatformBoth    = 3
)

const (
	ServiceID_ShareSDK = 0
	ServiceID_WEIXIN   = 1
)

type ServiceRecord struct {
	ID         uint32 `gorm:"primary_key"`
	CreatedAt  int64
	UpdatedAt  int64
	ServiceId  int
	PlatformId int
	Key        string
	Secret     string
}

func FindServiceInfo(serviceId int, platformId int) *ServiceRecord {
	u := new(ServiceRecord)
	if err := db.DB().
		Where(&ServiceRecord{ServiceId: serviceId, PlatformId: platformId}).
		Find(u).Error; err == nil {
		return u
	} else if err != gorm.ErrRecordNotFound {
		log.Logger().Error("error when FindServiceInfo err ", err.Error())
	}
	return nil
}
