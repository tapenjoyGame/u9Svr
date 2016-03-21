package channelPayNotify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/orm"
	"io"
	"net/http"
	"net/url"
	"time"
	"u9/api/common"
	"u9/models"
	"u9/tool"
)

type PayNotify interface {
	CheckUrlParam() (err error)
	ParseParam() (err error)
	CheckSign() (err error)
	Handle() (err error)
	ParseChannelRet() (err error)
	GetResult() (ret string)
}

const (
	err_noerror = iota
	err_checkUrlParam
	err_getPackageParam
	err_parseProductKey
	err_parseOrderRequest
	err_parseLoginRequest
	err_checkSign
	err_parseUrlParam
	err_handleOrder
	err_notifyProductSvr
	err_orderIsNotExist
	err_channelUserIsNotExist
	err_payAmountError
)

type Base struct {
	ProductRet common.BasicRet
	channelId  int
	productId  int
	urlParams  *url.Values
	urlKeys    *[]string

	loginRequest models.LoginRequest
	orderRequest models.OrderRequest
	payOrder     models.PayOrder

	orderId        string
	channelUserId  string
	channelOrderId string
	payAmount      int
	productKey     string

	callbackRet int //0:success 1:failure
}

func (this *Base) Init(channelId, productId int, urlParams *url.Values, urlKeys *[]string) {
	this.channelId = channelId

	this.productId = productId
	this.ProductRet.Init()

	this.urlParams = urlParams
	this.urlKeys = urlKeys

	this.callbackRet = err_noerror
}

func (this *Base) CheckUrlParam() (err error) {
	defer func() {
		if err != nil {
			this.callbackRet = err_checkUrlParam
			beego.Trace(err)
		}
	}()

	for _, urlKey := range *(this.urlKeys) {
		if this.urlParams.Get(urlKey) == "" {
			err = errors.New(fmt.Sprintf("Require %s", urlKey))
			return
		}
	}
	return
}

func (this *Base) getPackageParam(key string) (ret string, err error) {
	defer func() {
		if err != nil {
			this.callbackRet = err_getPackageParam
			beego.Trace(err)
		}
	}()

	var pp models.PackageParam
	err = pp.Query().Filter("channelId", this.channelId).Filter("productId", this.productId).One(&pp)
	if err != nil {
		return
	}

	args := new(map[string]string)
	if err = json.Unmarshal([]byte(pp.XmlParam), args); err != nil {
		return
	}

	ok := false
	if ret, ok = (*args)[key]; !ok {
		msg := fmt.Sprintf("PackageParam %s is empty.", key)
		err = errors.New(msg)
		return
	}
	return
}

func (this *Base) parseProductKey() (err error) {
	defer func() {
		if err != nil {
			this.callbackRet = err_parseProductKey
			beego.Trace(err)
		}
	}()

	p := new(models.Product)
	if err = p.Query().Filter("Id", this.productId).One(p); err != nil {
		return
	}

	if p.AppKey == "" {
		err = errors.New("Product's appkey is empty.")
		return
	} else {
		this.productKey = p.AppKey
	}
	return
}

func (this *Base) parseOrderRequest() (err error) {
	defer func() {
		if err != nil {
			this.callbackRet = err_parseOrderRequest
			beego.Trace(err)
		}
	}()

	orqs := this.orderRequest.Query().Filter("ChannelId", this.channelId).
		Filter("ProductId", this.productId).Filter("OrderId", this.orderId)
	if orqs.Exist() == false {
		format := "orderRequest is not exist,channelId:%d,productId:%d,orderId:%s"
		err = errors.New(fmt.Sprintf(format, this.channelId, this.productId, this.orderId))
		return
	}
	if err = orqs.One(&this.orderRequest); err != nil {
		return
	}

	var jsonBytes []byte
	if jsonBytes, err = json.Marshal(this.urlParams); err != nil {
		return
	}

	this.orderRequest.ChannelLog = string(jsonBytes)
	if err = this.orderRequest.Update("ChannelLog"); err != nil {
		return
	}
	return
}

func (this *Base) parseLoginRequest() (err error) {
	err = this.loginRequest.Query().Filter("Userid", this.orderRequest.UserId).One(&this.loginRequest)
	if err != nil {
		this.callbackRet = err_parseLoginRequest
		beego.Trace(err)
		return
	}
	return
}

func (this *Base) ParseChannelRet() (err error) {
	if this.orderId != this.orderRequest.OrderId {
		this.callbackRet = err_orderIsNotExist
		msg := fmt.Sprintf("order is invalid, db(%s), url(%s)", this.orderRequest.OrderId, this.orderId)
		beego.Error(msg)
		err = errors.New(msg)
		return
	}

	if this.orderRequest.ReqAmount != this.payAmount {
		this.callbackRet = err_payAmountError
		msg := fmt.Sprintf("payAmount is invalid, db(%d), url(%d)", this.orderRequest.ReqAmount, this.payAmount)
		beego.Error(msg)
		err = errors.New(msg)
		return
	}

	if this.channelUserId != this.loginRequest.ChannelUserid {
		this.callbackRet = err_channelUserIsNotExist
		msg := fmt.Sprintf("channelUserId is invalid, db(%s), url(%s)", this.loginRequest.ChannelUserid, this.channelUserId)
		beego.Error(msg)
		err = errors.New(msg)
		return
	}
	return
}

func (this *Base) CheckSign() (err error) {
	this.callbackRet = err_checkSign
	return
}

func (this *Base) parseUrlParam() (err error) {
	this.callbackRet = err_parseUrlParam
	return
}

func (this *Base) ParseParam() (err error) {
	if err = this.parseOrderRequest(); err != nil {
		return
	}
	if err = this.parseLoginRequest(); err != nil {
		return
	}
	if err = this.parseProductKey(); err != nil {
		return
	}
	return
}

func (this *Base) notifyProductSvr() (err error) {
	defer func() {
		if err != nil {
			this.callbackRet = err_notifyProductSvr
			beego.Trace(err)
		}
	}()

	signContext := fmt.Sprintf("%s%s%s%s",
		this.orderRequest.ProductOrderId, this.orderId, this.channelOrderId, this.productKey)

	if this.orderRequest.CallbackUrl != "" {
		format := this.orderRequest.CallbackUrl + "/?" +
			"UserId=%s" + "&" + "OrderId=%s" + "&" + "ChannelId=%d" + "&" + "ChannelOrderId=%s" + "&" +
			"ChannelUserId=%s" + "&" + "ProductId=%d" + "&" + "ProductOrderId=%s" + "&" +
			"ReqAmount=%d" + "&" + "PayAmount=%d" + "&" + "Sign=%s" + "&" + "Code=%d"
		notifyUrl := fmt.Sprintf(format,
			this.orderRequest.UserId, this.orderId, this.channelId, this.channelOrderId,
			this.channelUserId, this.productId, this.orderRequest.ProductOrderId,
			this.orderRequest.ReqAmount, this.payAmount, tool.Md5([]byte(signContext)), this.callbackRet)

		req := httplib.Get(notifyUrl)
		if _, err = req.Response(); err != nil {
			beego.Trace(notifyUrl)
			return
		}
		err = req.ToJSON(&this.ProductRet)
		if err != nil {
			beego.Trace(notifyUrl)
			return
		}
	} else {
		this.ProductRet.SetCode(0)
	}

	this.orderRequest.ProductCode = this.ProductRet.Code
	this.orderRequest.ProductMessage = this.ProductRet.Message
	if this.ProductRet.Code == 0 {
		this.orderRequest.State = 2
	}
	this.orderRequest.Update("State", "ProductMessage", "ProductCode")
	return
}

func (this *Base) request(url string) (ret string, err error) {
	req := httplib.Get(url)
	var resp *http.Response
	resp, err = req.Response()
	if resp, err = req.Response(); err != nil {
		return
	}

	var buffer bytes.Buffer
	if _, err = io.Copy(&buffer, resp.Body); err != nil {
		return
	}
	bytes := buffer.Bytes()
	return string(bytes), nil
}

func (this *Base) handleOrder() (err error) {
	defer func() {
		if err != nil {
			this.callbackRet = err_handleOrder
			//beego.Error(err)
		}
	}()

	this.orderRequest.State = 1
	err = this.orderRequest.Update("State")
	if err != nil {
		return
	}

	created := false
	this.payOrder = models.PayOrder{OrderId: this.orderId, ChannelOrderId: this.channelOrderId}
	if created, _, err = orm.NewOrm().ReadOrCreate(&this.payOrder, "OrderId", "ChannelOrderId"); err != nil {
		beego.Error(err, ":orderId(", this.orderId, ")channelOrderId(", this.channelOrderId, ")")
		return
	}

	if !created {
		//err = errors.New("order is exist")
		return
	}

	this.payOrder.PayAmount = this.payAmount
	this.payOrder.PayTime = time.Now()
	if err = this.payOrder.Update("PayAmount"); err != nil {
		beego.Error(err, ":payAmount(", this.payAmount, ")")
		return
	}
	return
}

func (this *Base) Handle() (err error) {
	switch this.orderRequest.State {
	case 0, 1:
		if err = this.handleOrder(); err != nil {
			return err
		}
		if err = this.notifyProductSvr(); err != nil {
			return err
		}
	case 2:
		return nil
	default:
		return errors.New("It isn't exist orderRequest state")
	}
	return
}