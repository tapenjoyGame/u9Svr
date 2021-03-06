package lrBefore

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"net/url"
)

type qihoo360ChannelRet struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	Sex        string `json:"sex"`
	Area       string `json:"area"`
	Nick       string `json:"nick"`
	Error_code string `json:"error_code"`
	Error      string `json:"error"`
}

type Qihoo360 struct {
	base
}

func NewQihoo360() *Qihoo360 {
	ret := new(Qihoo360)
	return ret
}

func (this *Qihoo360) Init(param *Param) (err error) {
	this.base.Init(param)
	this.channelRet = new(qihoo360ChannelRet)
	return
}

func (this *Qihoo360) Exec() (ret string, err error) {
	this.IsHttps = true
	this.Method = "GET"

	token := url.QueryEscape(this.param.Token)
	format := "?access_token=%s"
	this.Url = "https://openapi.360.cn/user/me" + fmt.Sprintf(format, token)
	beego.Trace(this.Url)

	this.base.InitParam()

	if err = this.GetResponse(); err != nil {
		beego.Error(err)
		return
	}

	beego.Trace("result:", this.Result)
	if err = json.Unmarshal([]byte(this.Result), &this.channelRet); err != nil {
		beego.Error(err)
		return
	}

	channelret := this.channelRet.(*qihoo360ChannelRet)
	if channelret.Error_code != "" {
		err = errors.New(this.Result)
		beego.Error(fmt.Sprintf("channelRet:%+v", this.channelRet))
		return
	}

	this.param.ChannelUserId = channelret.Id
	this.param.ChannelUserName = channelret.Name

	ret = this.Result
	return
}
