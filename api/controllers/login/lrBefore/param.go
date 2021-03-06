package lrBefore

import (
	"github.com/astaxie/beego/validation"
	"u9/models"
)

type MobileInfo struct {
	Imei string `json:"Imei"`
	Imsi string `json:"Imsi"`
}

type Param struct {
	ProductId       int    `json:"ProductId"`
	ChannelId       int    `json:"ChannelId"`
	ChannelUserId   string `json:"ChannelUserId"`
	ChannelUserName string `json:"ChannelUserName"`
	Token           string `json:"Token"`
	MobileInfo      string `json:"MobileInfo"`
	Ext             string `json:"Ext"`
	IsDebug         bool   `json:"IsDebug"`
}

func (this *Param) Valid(v *validation.Validation) {
	switch {
	case this.ChannelId <= 0:
		v.SetError("1001", "Require channelId")
		return
	case this.Token == "":
		v.SetError("1003", "Require token")
		return
	case this.ProductId <= 0:
		v.SetError("1002", "Require productId")
		return
	}

	switch {
	case new(models.Channel).Query().Filter("id", this.ChannelId).Exist() == false:
		v.SetError("1001", "Channel is not exist in database")
	case new(models.Product).Query().Filter("id", this.ProductId).Exist() == false:
		v.SetError("1002", "Product is not exist in database")
	}
}
