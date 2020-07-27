package GoProxy

import (
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var lock sync.Mutex
var ProxyUrls ProxySwitcher
var IpChan chan IP
var GetIpType = GetEnv("GETIPTYPE", "PollingProxyPool")
var log = logrus.New()

type ProxySwitcher struct {
	ProxyURLs []string
	index     uint32
	size      int32
}

type IP struct {
	Proxy string
	First bool
}

func (r *ProxySwitcher) GetRoundProxy() string {
	if len(r.ProxyURLs) > 0 {
		u := r.ProxyURLs[r.index%uint32(len(r.ProxyURLs))]
		atomic.AddUint32(&r.index, 1)
		return u
	}
	return ""
}

func CronCheckProxy() {
	for {
		time.Sleep(30 * time.Second)
		log.Infoln("[Proxy] 当前拥有的 IP 数量为:", len(ProxyUrls.ProxyURLs))
		lock.Lock()
		urls := ProxyUrls.ProxyURLs
		ProxyUrls.ProxyURLs = []string{}
		lock.Unlock()
		for _, url := range urls {
			ip := IP{Proxy: url, First: false}
			IpChan <- ip
		}
	}
}

func init() {
	ProxyUrls = ProxySwitcher{}
	IpChan = make(chan IP, 100)
}

func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
