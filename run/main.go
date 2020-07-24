package main

import "GoProxy"

func main() {
	defer close(GoProxy.IpChan)
	go GoProxy.CronCheckProxy()
	go GoProxy.Getter()
	go GoProxy.Tester()
	GoProxy.ProxySever()
}
