package GoProxy

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
)

func ProxySever() {
	l, err := net.Listen("tcp", ":3128")
	if err != nil {
		log.Panic(err)
	}

	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handle(client)
	}
}

func handle(client net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			debug.PrintStack()
		}
	}()
	if client == nil {
		return
	}
	log.Printf("[Proxy] client tcp tunnel connection: [%v] -> [%v]", client.LocalAddr().String(), client.RemoteAddr().String())
	// client.SetDeadline(time.Now().Add(time.Duration(10) * time.Second))
	defer client.Close()

	var b [1024]byte
	n, err := client.Read(b[:])
	if err != nil || bytes.IndexByte(b[:], '\n') == -1 {
		log.Errorln("[Proxy] 读取应用层的所有数据出错了:", err, b)
		return
	}
	var method, host, address string
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)
	hostPortURL, err := url.Parse(host)
	if err != nil {
		log.Println(err)
		return
	}
	if hostPortURL.Opaque == "443" {
		address = hostPortURL.Scheme + ":443"
	} else {
		if strings.Index(hostPortURL.Host, ":") == -1 {
			address = hostPortURL.Host + ":80"
		} else {
			address = hostPortURL.Host
		}
	}

	server, err := proxyDial("tcp", address)
	if err != nil {
		log.Errorln("[Proxy] 建立连接失败:", err)
		return
	}
	//在应用层完成数据转发后，关闭传输层的通道
	defer server.Close()
	log.Infof("[Proxy] server tcp tunnel connection: [%v] -> [%v]", server.LocalAddr().String(), server.RemoteAddr().String())
	// server.SetDeadline(time.Now().Add(time.Duration(10) * time.Second))

	if method == "CONNECT" {
		fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
	} else {
		log.Infoln("[Proxy] server write", method) //其它协议
		server.Write(b[:n])
	}

	go func() {
		io.Copy(server, client)
	}()
	io.Copy(client, server) //阻塞转发
}

func proxyDial(network, addr string) (net.Conn, error) {
	c, err := func() (net.Conn, error) {
		u, _ := url.Parse(ProxyUrls.GetRoundProxy())
		c, err := net.DialTimeout("tcp", u.Host, time.Second*5)
		if err != nil {
			return nil, err
		}

		reqURL, err := url.Parse("http://" + addr)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest(http.MethodConnect, reqURL.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Close = false
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.3")

		err = req.Write(c)
		if err != nil {
			return nil, err
		}

		resp, err := http.ReadResponse(bufio.NewReader(c), req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		log.Println(resp.StatusCode, resp.Status, resp.Proto, resp.Header)
		if resp.StatusCode != 200 {
			err = fmt.Errorf("[Proxy] Connect server using proxy error, StatusCode [%d]", resp.StatusCode)
			return nil, err
		}
		return c, err
	}()
	if c == nil || err != nil { //代理异常
		log.Errorln("[Proxy] 代理异常：", c, err)
		log.Infoln("[Proxy] 本地直接转发：", c, err)
		return net.Dial(network, addr)
	}
	log.Infof("[Proxy] 代理正常,tunnel 信息 [%v] -> [%v]", c.LocalAddr().String(), c.RemoteAddr().String())
	return c, err
}
