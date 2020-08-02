# Go-Proxy
简单的基于 Go 的 HTTP(S) 二级代理池，包括抓取、测试 IP，绑定端口随机转发一个 IP
## 启动
```bash
# uninx shell 启动方法 windows 可以直接编译出 exe 文件运行
cd run
go build main.go
./main
```
支持 docker 启动
```bash
docker build -t proxy:latest .
docker run -d -p 3128:3128 --name proxy proxy:latest
```
## 使用方法
直接绑定 3128 端口即可，会动态转发一个随机 IP

代理转发目前的规则为轮询
```go
package main

import (
    "github.com/guonaihong/gout"
	"net/http"
)

func main(){
c := &http.Client{}
_ := gout.New(c).GET("https://www.baidu.com").
		Debug(false).
		SetProxy("http://localhost:3128").
		SetHeader(gout.H{
			"accept":     "*/*",
			"user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_16_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36 Edg/84.0.522.40",
		}).
		Code(&code).
		BindBody(&resp).
		SetTimeout(10 * time.Second).
		Filter().Retry().Attempt(5).WaitTime(500 * time.Millisecond).
		Do()
}
```

### 更改 ip 获取
修改 getter.go 文件里的 GetIp** 的方法即可，目前有两个示例，一个来自 `猿人云 IP`
一个是 `芝麻 HTTP`(都是付费代理)，是两种不同的获取 API 方式，需要更改 init.go 文件里的 GetIpType 
的值来选择启用哪种,也可以设置环境变量 GETIPTYPE 来使用。


#####TODO:
目前我同一时间只会使用一家的服务所以就这么设计了，之后也许会改

