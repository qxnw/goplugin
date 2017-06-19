package main

import "github.com/qxnw/goplugin"

var reg *goplugin.Registry

func init() {
	reg = goplugin.NewRegistry()
	reg.Register("/100bm/send/sms", NewSendSms())
}

func Register(name string, handler goplugin.Handler) {
	reg.Register(name, handler)
}
func GetServices() []string {
	return reg.Services
}
func GetHandlers() map[string]goplugin.Handler {
	return reg.ServiceHandlers
}
