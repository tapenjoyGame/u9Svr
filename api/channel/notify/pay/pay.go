package channelPayNotify

import (
	"errors"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"net/url"
)

func CallPayNotify(channelId, productId int, urlParams *url.Values, ctx *context.Context) (ret string, err error) {
	beego.Trace("CallPayNotify")
	var pn PayNotify

	defer func() {
		if pn != nil {
			ret = pn.GetResult()
		} else {
			ret = "failure"
		}
		beego.Trace("GetResult:", ret)
	}()

	switch channelId {
	case 100:
		pn = NewTest(channelId, productId, urlParams)
	case 101:
		pn = NewDangle(channelId, productId, urlParams)
	case 103:
		pn = NewAnZhi(channelId, productId, urlParams)
	case 104:
		pn = NewCCPay(channelId, productId, urlParams)
	case 106:
		pn = NewGFan(channelId, productId, urlParams, ctx)
	case 107:
		pn = NewGuopan(channelId, productId, urlParams)
	case 109:
		pn = NewM4399(channelId, productId, urlParams)
	case 110:
		pn = NewOppo(channelId, productId, urlParams)
	case 112:
		pn = NewMeizu(channelId, productId, urlParams)
	case 117:
		pn = NewWandoujia(channelId, productId, urlParams)
	case 118:
		pn = NewXiaomi(channelId, productId, urlParams)
	case 120:
		pn = NewAmigo(channelId, productId, urlParams)
	case 122:
		pn = NewCYGame(channelId, productId, urlParams)
	case 123:
		pn = NewXmw(channelId, productId, urlParams)
	case 125:
		pn = NewHaimawan(channelId, productId, urlParams)
	case 126:
		pn = NewLeTV(channelId, productId, urlParams)
	case 127:
		pn = NewHTC(channelId, productId, urlParams, ctx)
	case 128:
		pn = NewZhuoYi(channelId, productId, urlParams)
	case 129:
		pn = NewShouMeng(channelId, productId, urlParams)
	case 130:
		pn = NewYYH(channelId, productId, urlParams)
	default:
		err = errors.New("channelId isn't exist.")
		return
	}
	beego.Trace("CheckUrlParam")
	if err = pn.CheckUrlParam(); err != nil {
		return
	}
	beego.Trace("ParseParam")
	if err = pn.ParseParam(); err != nil {
		return
	}
	beego.Trace("CheckSign")
	if err = pn.CheckSign(); err != nil {
		return
	}
	beego.Trace("Handle")
	if err = pn.Handle(); err != nil {
		return
	}
	beego.Trace("ParseChannelRet")
	if err = pn.ParseChannelRet(); err != nil {
		return
	}
	return
}
