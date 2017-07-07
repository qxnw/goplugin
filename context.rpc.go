package goplugin

import (
	"fmt"

	"github.com/qxnw/lib4go/jsons"
)

//ContextRPC MQ操作实例
type ContextRPC struct {
	ctx *PluginContext
}

//Reset 重置context
func (cr *ContextRPC) Reset(ctx *PluginContext) {
	cr.ctx = ctx
}

//RequestFailRetry RPC请求
func (cr *ContextRPC) RequestFailRetry(service string, input map[string]string, times int) (status int, r string, param map[string]string, err error) {
	status, r, param, err = cr.ctx.rpc.RequestFailRetry(service, input, times)
	if err != nil || status != 200 {
		err = fmt.Errorf("rpc请求(%s)失败:%d,err:%v", service, status, err)
		return
	}
	return
}

//Request RPC请求
func (cr *ContextRPC) Request(service string, input map[string]string, failFast bool) (status int, r string, param map[string]string, err error) {
	status, r, param, err = cr.ctx.rpc.Request(service, input, failFast)
	if err != nil || status != 200 {
		err = fmt.Errorf("rpc请求(%s)失败:%d,err:%v", service, status, err)
		return
	}
	return
}

//RequestMap RPC请求返回结果转换为map
func (cr *ContextRPC) RequestMap(service string, input map[string]string, failFast bool) (status int, r map[string]interface{}, param map[string]string, err error) {
	status, result, _, err := cr.Request(service, input, failFast)
	if err != nil {
		return
	}
	r, err = jsons.Unmarshal([]byte(result))
	if err != nil {
		err = fmt.Errorf("rpc请求返结果不是有效的json串:%s,%v,%s,err:%v", service, input, result, err)
		return
	}
	return
}
