package loginRequest

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"net/url"
	"u9/models"
	"u9/tool"
)

type ZhuoyiChannelRet struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Zhuoyi struct {
	Lr
	channelRet ZhuoyiChannelRet
}

func LrNewZhuoYi(mlr *models.LoginRequest, args *map[string]interface{}) *Zhuoyi {
	ret := new(Zhuoyi)
	ret.Init(mlr, args)
	return ret
}

func (this *Zhuoyi) Init(mlr *models.LoginRequest, args *map[string]interface{}) {
	this.Lr.Init(mlr)
	appid := (*args)["zy_app_id"].(string)
	serverKey := (*args)["zy_app_secret"].(string)

	format := "uid=%s&access_token=%s&app_id=%s&key=%s"
	content := fmt.Sprintf(format, this.mlr.ChannelUserid, this.mlr.Token, appid, serverKey)
	sign := tool.Md5([]byte(content))
	format = "http://open.zhuoyi.com/phone/index.php/ILoginAuth/auth?access_token=%s&uid=%s&app_id=%s&sign=%s"
	this.Url = fmt.Sprintf(format, this.mlr.Token, this.mlr.ChannelUserid, appid, sign)
	beego.Trace(this.Url)
}

func (this *Zhuoyi) ParseChannelRet() (err error) {
	err = json.Unmarshal([]byte(this.Result), &this.channelRet)
	message, _ := url.QueryUnescape(this.channelRet.Message)
	beego.Trace(this.Result, ":", "message:", message)
	return
}

func (this *Zhuoyi) CheckChannelRet() bool {
	return this.channelRet.Code == "0"
}
