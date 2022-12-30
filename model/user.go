package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type User1 struct {
	gorm.Model
	Name   string `gorm:"size:255;not null;" json:"name"`
	Email  string `gorm:"size:100;not null;unique" json:"email"`
	Mobile string `gorm:"size:100;not null;" json:"mobile"`
	//Transactions []Transaction.Transaction `gorm:"ForeignKey:UserID"`
}

type User struct {
	UserID           string    `gorm:"column:user_id;primary_key;size:64" json:"userID,omitempty"`
	Nickname         string    `gorm:"column:name;size:255" json:"nickname,omitempty"`
	FaceURL          string    `gorm:"column:face_url;size:255" json:"faceURL,omitempty"`
	Gender           int32     `gorm:"column:gender" json:"gender,omitempty"`
	PhoneNumber      string    `gorm:"column:phone_number;size:32" json:"phoneNumber,omitempty"`
	Birth            time.Time `gorm:"column:birth" json:"birth,omitempty"`
	Email            string    `gorm:"column:email;size:64" json:"email,omitempty"`
	Ex               string    `gorm:"column:ex;size:1024" json:"ex,omitempty"`
	CreateTime       time.Time `gorm:"column:create_time;index:create_time" json:"createTime,omitempty"`
	AppMangerLevel   int32     `gorm:"column:app_manger_level" json:"appMangerLevel,omitempty"`
	GlobalRecvMsgOpt int32     `gorm:"column:global_recv_msg_opt" json:"global_recv_msg_opt,omitempty"`

	status int32 `gorm:"column:status" json:"status,omitempty"`
}
