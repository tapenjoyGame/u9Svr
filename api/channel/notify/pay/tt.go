package channelPayNotify

import (
	"crypto/rsa"
	"encoding/json"
	// "errors"
	"bytes"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"io"
	"net/url"
	"strings"
	"u9/tool"
)

var ttUrlKeys []string = []string{"order", "sign", "sign_type"}

const (
	err_ttParsePayKey      = 12701
	err_ttResultFailure    = 12702
	err_ttInitRsaPublicKey = 12703
	err_ttParseBody        = 12704
)

//tt
type TT struct {
	Base
	payKey     string
	response   Response
	tt_result TT_Result
	ctx        *context.Context
}
type Response struct {
	Order     string
	Sign      string
	Sign_type string
}
type TT_Result struct {
	Uid   		 string     `json:"uid"`
	GameId     	 string 	`json:"gameId"`
	SDKOrderId   string     `json:"sdkOrderId"`
	CpOrderId    string 	`json:"cpOrderId"`
	PayFee     	 string 	`json:"payFee"`
	PayResult 	 string 	`json:"payResult"`
	PayDate 	 string 	`json:"payDate"`
	ExInfo   	 string 	`json:"exInfo"`
}

var (
	ttRsaPublicKey *rsa.PublicKey
)

func NewTT(channelId, productId int, urlParams *url.Values, ctx *context.Context) *TT {
	ret := new(TT)
	ret.Init(channelId, productId, urlParams, ctx)
	return ret
}

func (this *TT) Init(channelId, productId int, urlParams *url.Values, ctx *context.Context) {
	this.Base.Init(channelId, productId, urlParams, &ttUrlKeys)
	this.ctx = ctx
}

func (this *TT) parsePayKey() (err error) {
	defer func() {
		if err != nil {
			this.callbackRet = err_ttParsePayKey
			beego.Trace(err)
		}
	}()
	this.payKey, err = this.getPackageParam("TT_SDK_PUBLICKEY")
	return
}

func (this *TT) CheckUrlParam() (err error) {
	return
}

func (this *TT) parseUrlParam() (err error) {
	defer func() {
		if err != nil {
			this.callbackRet = err_parseUrlParam
			beego.Trace(err)
		}
	}()
	// this.order = url.QueryEscape(this.response.Order)

	beego.Trace(this.response.Order)
	json.Unmarshal([]byte(this.response.Order), &this.tt_result)
	beego.Trace(this.tt_result)
	this.orderId = this.tt_result.Game_order_id
	this.channelOrderId = this.tt_result.Jolo_order_id
	this.payAmount = this.tt_result.Real_amount

	return
}

func (this *TT) ParseChannelRet() (err error) {
	if result := this.tt_result.Result_code; result != 1 {
		this.callbackRet = err_ttResultFailure
	}
	return
}
func (this *TT) parseBody() (err error) {
	defer func() {
		if err != nil {
			this.callbackRet = err_ttParseBody
			beego.Trace(err)
		}
	}()

	var buffer bytes.Buffer
	if _, err = io.Copy(&buffer, this.ctx.Request.Body); err != nil {
		return
	}
	content := string(buffer.Bytes())
	var newValues url.Values
	if newValues, err = url.ParseQuery(content); err != nil {
		return
	}
	this.response.Order = newValues.Get("order")
	this.response.Sign = newValues.Get("sign")
	this.response.Sign_type = newValues.Get("sign_type")
	this.response.Sign = strings.Replace(this.response.Sign, "\"", "", 2)
	beego.Trace(this.response.Order)
	beego.Trace(this.response.Sign)
	beego.Trace(this.response.Sign_type)
	return
}

func (this *TT) ParseParam() (err error) {
	if err = this.parseBody(); err != nil {
		return
	}
	if err = this.parseUrlParam(); err != nil {
		return
	}
	if err = this.parsePayKey(); err != nil {
		return
	}
	if err = this.Base.ParseParam(); err != nil {
		return
	}
	this.channelUserId = this.loginRequest.ChannelUserid
	this.initRsaPublicKey()
	return
}

func (this *TT) initRsaPublicKey() (err error) {
	defer func() {
		if err != nil {
			this.callbackRet = err_ttInitRsaPublicKey
			beego.Trace(err)
		}
	}()

	if ttRsaPublicKey == nil {
		ttRsaPublicKey, err = tool.ParsePKIXPublicKeyWithStr(this.payKey)
		if err != nil {
			beego.Error(err)
			return err
		}
	}
	// beego.Trace(TTRsaPublicKey)
	return nil
}

func (this *TT) CheckSign() (err error) {
	defer func() {
		if err != nil {
			this.callbackRet = err_checkSign
			beego.Trace(err)
		}
	}()

	// if sign := tool.RsaVerifyPKCS1v15(TTRsaPublicKey, this.order); sign != this.urlParams.Get("signature") {
	// 	msg := fmt.Sprintf("Sign is invalid, context:%s, sign:%s", context, sign)
	// 	err = errors.New(msg)
	// 	return
	// }
	if err = tool.RsaVerifyPKCS1v15(ttRsaPublicKey, this.response.Order, this.response.Sign); err != nil {
		msg := fmt.Sprintf("RsaVerifyPK CS1v15 exception: context:%s, sign:%s", this.response.Order, this.response.Sign)
		beego.Trace(msg)
		return err
	}
	return
}

func (this *TT) GetResult() (ret string) {
	if this.callbackRet == err_noerror {
		ret = `{"head":{"result":"0","message":"成功"}}`
	} else {
		ret = `{"head":{"result":"1","message":"失败"}}`
	}
	return
}

/*
  signature rule: md5("order=xxxx&money=xxxx&mid=xxxx&time=xxxx&result=x&ext=xxx&key=xxxx")
  test url:
  http://192.168.0.185/api/channelPayNotify/1000/101/?
  order=test20160116172500359&
  money=100.00&
  mid=test10086001&
  time=20160116172500&
  result=1&
  ext=game20160116175128772&
  signature=8f00a109716e819bfe0afb695c1addf1
*/
