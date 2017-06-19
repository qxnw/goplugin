package main

//go build -buildmode=plugin
import (
	"fmt"

	"github.com/qxnw/goplugin"
)

type cbsCore struct {
}

//GetServices 获取当前插件提供的所有服务
func (p *cbsCore) GetServices() []string {
	return GetServices()
}

//Handle 业务处理
func (p *cbsCore) Handle(name string, mode string, service string, c goplugin.Context, invoker goplugin.RPCInvoker) (status int, result interface{}, param map[string]interface{}, err error) {
	if h, ok := GetHandlers()[service]; ok {
		status, r, err := h.Handle(service, c, invoker)
		if err != nil || status != 200 {
			return status, result, nil, fmt.Errorf("send.sms:%v", err)
		}
		return status, r, nil, err

	}
	return 404, "", nil, fmt.Errorf("send.sms 未找到服务:%s", service)
}

//GetWorker 获取当前worker
func GetWorker() goplugin.Worker {
	return &cbsCore{}
}
