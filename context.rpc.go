package goplugin

import (
	"fmt"

	"github.com/qxnw/lib4go/jsons"
)

//RPCRquest RPC请求
func (w *PluginContext) RPCRquest(service string, input map[string]string, failFast bool) (status int, r string, param map[string]string, err error) {
	status, r, param, err = w.RPC.Request(service, input, failFast)
	return
}

//RPCRquestMap RPC请求返回结果转换为map
func (w *PluginContext) RPCRquestMap(service string, input map[string]string, failFast bool) (status int, r map[string]interface{}, param map[string]string, err error) {
	status, rx, param, err := w.RPC.Request(service, input, failFast)
	if err != nil || status != 200 {
		err = fmt.Errorf("rpc请求(%s)失败:%d,err:%v", service, status, err)
		return
	}
	r, err = jsons.Unmarshal([]byte(rx))
	if err != nil {
		err = fmt.Errorf("rpc请求(%s)返回结果解析失败:%s,err:%v", service, rx, err)
		return
	}
	return
}
