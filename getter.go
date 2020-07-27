package GoProxy

import (
	"fmt"
	"github.com/guonaihong/gout"
	"github.com/guonaihong/gout/dataflow"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

type ApeIPres struct {
	Msg  string      `json:"msg"`
	Code int         `json:"code"`
	Data []ApeIPData `json:"data"`
}

type ZhiMaIPres struct {
	Msg     string        `json:"msg"`
	Code    int           `json:"code"`
	Success bool          `json:"success"`
	Data    []ZhiMaIPData `json:"data"`
}

type ApeIPData struct {
	OriginIp string `json:"origin_ip"`
	Ip       string `json:"ip"`
	Port     int    `json:"port"`
	Address  string `json:"address"`
	Province string `json:"province"`
	City     string `json:"city"`
	Isp      string `json:"isp"`
	ExpireAt string `json:"expire_at"`
}

type ZhiMaIPData struct {
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	ExpireTime string `json:"expire_time"`
	City       string `json:"city"`
	Isp        string `json:"isp"`
}

func Getter() {
	switch {
	case GetIpType == "PollingProxyPool":
		for {
			if ProxyUrls.size <= int32(20) {
				GetIpFromZhiMa()
			}
			time.Sleep(2 * time.Second)
		}
	case GetIpType == "PollingGetApi":
		for {
			GetIpFromApeYun()
			time.Sleep(10 * time.Second)
		}
	default:
		log.Fatalln("[Getter] 不受支持的 GetIpType :", GetIpType)
	}
}

func GetIpFromApeYun() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorln("[Getter] 获取器出错了:", err)
			debug.PrintStack()
		}
	}()
	var (
		name   = "猿人云"
		resp   ApeIPres
		errStr string
		c      = &http.Client{}
	)

	err := gout.New(c).GET("http://tunnel-api.apeyun.com/q").
		Debug(false).
		SetQuery(gout.H{
			"id":        "",
			"secret":    "",
			"limit":     "5",
			"format":    "json",
			"auth_mode": "hand",
		}).
		SetHeader(gout.H{
			"accept":     "*/*",
			"user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_16_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36 Edg/84.0.522.40",
		}).
		Callback(func(c *dataflow.Context) error {
			switch c.Code {
			case 200: //http code为200时，服务端返回的是json 结构
				c.BindJSON(&resp)
			default: //http code为404时，服务端返回是html 字符串
				c.BindBody(&errStr)
			}
			return nil
		}).
		Filter().Retry().Attempt(3).WaitTime(time.Second).MaxWaitTime(6 * time.Second).
		Do()
	if err != nil {
		log.Errorf("[Getter] %v 获取 IP 出错了: %v", name, err)
	}
	if resp.Code == 200 && resp.Msg == "success" {
		for _, d := range resp.Data {
			proxyUrl := fmt.Sprintf(`http://%v`, d.Ip+":"+strconv.Itoa(d.Port))
			log.Infof("[Getter] %v 获取到 IP: [%v]", name, proxyUrl)
			ip := IP{Proxy: proxyUrl, First: true}
			IpChan <- ip
		}
	} else {
		log.Errorf("[Getter] %v 获取 IP 出错了: %v", name, errStr)
	}
}

func GetIpFromZhiMa() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorln("[Getter] 获取器出错了:", err)
			debug.PrintStack()
		}
	}()

	var (
		name   = "芝麻 HTTP"
		resp   ZhiMaIPres
		errStr string
		c      = &http.Client{}
	)

	err := gout.New(c).GET("http://webapi.http.zhimacangku.com/getip").
		Debug(false).
		SetQuery(gout.H{
			"num":     "20",
			"type":    "2",
			"pro":     "0",
			"city":    "0",
			"yys":     "0",
			"port":    "11",
			"pack":    "",
			"ts":      "1",
			"ys":      "0",
			"cs":      "0",
			"lb":      "1",
			"sb":      "0",
			"pb":      "4",
			"mr":      "2",
			"regions": "",
		}).
		SetHeader(gout.H{
			"accept":     "*/*",
			"user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_16_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36 Edg/84.0.522.40",
		}).
		Callback(func(c *dataflow.Context) error {
			switch c.Code {
			case 200: //http code为200时，服务端返回的是json 结构
				c.BindJSON(&resp)
			default:
				c.BindBody(&errStr)
			}
			return nil
		}).
		Filter().Retry().Attempt(3).WaitTime(time.Second).MaxWaitTime(6 * time.Second).
		Do()
	if err != nil {
		log.Errorf("[Getter] %v 获取 IP 出错了: %v", name, errStr)
	}
	if resp.Code == 0 && resp.Success == true {
		for _, d := range resp.Data {
			proxyUrl := fmt.Sprintf(`http://%v`, d.Ip+":"+strconv.Itoa(d.Port))
			log.Infof("[Getter] %v 获取到 IP: [%v]", name, proxyUrl)
			ip := IP{Proxy: proxyUrl, First: true}
			IpChan <- ip
		}
	} else {
		log.Errorf("[Getter] %v 获取 IP 出错了: %v", name, errStr)
	}
}
