package GoProxy

import (
	"github.com/guonaihong/gout"
	"net/http"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"
)

func Tester() {
	for v := range IpChan {
		ip := v
		if (ip != IP{}) {
			go TestIP(ip)
		}
	}
}

func TestIP(ip IP) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorln("[Tester] 测试器出错了:", err)
			debug.PrintStack()
		}
	}()
	var (
		code int
		resp string
		c    = &http.Client{}
	)
	err := gout.New(c).GET("https://www.baidu.com").
		Debug(false).
		SetProxy(ip.Proxy).
		SetHeader(gout.H{
			"accept":     "*/*",
			"user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_16_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36 Edg/84.0.522.40",
		}).
		Code(&code).
		BindBody(&resp).
		SetTimeout(10 * time.Second).
		Filter().Retry().Attempt(5).WaitTime(500 * time.Millisecond).
		Do()
	if err != nil && strings.Contains(err.Error(), "retry fail") {
		log.Errorf("[Tester] 测试器测试 IP [%v] 出错了: %v", ip.Proxy, err)
		if ip.First == false {
			atomic.AddInt32(&ProxyUrls.size, -1)
		}
	} else {
		if ip.First == true {
			atomic.AddInt32(&ProxyUrls.size, 1)
		}
		lock.Lock()
		//log.Infof("[Tester] 测试器测试 IP [%v] 可用", proxyUrl)
		ProxyUrls.ProxyURLs = append(ProxyUrls.ProxyURLs, ip.Proxy)
		lock.Unlock()
	}
}
