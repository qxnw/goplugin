package goplugin

//编译指令
//go build -buildmode=plugin
import (
	"encoding/json"
	"fmt"
)

type wxBaseCore struct {
}

var registry *Registry

func init() {
	registry = NewRegistry()
}

//GetServices 获取当前插件提供的所有服务
func (p *wxBaseCore) GetServices() []string {
	return registry.Services
}

//Handle 业务处理
func (p *wxBaseCore) Handle(name string, mode string, service string, c Context, invoker RPCInvoker) (status int, result interface{}, param map[string]interface{}, err error) {
	if h, ok := registry.ServiceHandlers[service]; ok {
		status, r, err := h.Handle(service, c, invoker)
		if err != nil || status != 200 {
			return status, result, nil, fmt.Errorf("wx_base_core_api:%v", err)
		}
		buffer, err := json.Marshal(r)
		if err != nil {
			return 500, "", nil, fmt.Errorf("wx_base_core_api:输换输出结果出错%v", err)
		}
		result = string(buffer)
		return 200, result, nil, err

	}
	return 404, "", nil, fmt.Errorf("wx_base_core_api 未找到服务:%s", service)
}

//GetWorker 获取当前worker
func GetWorker() PluginWorker {
	return &wxBaseCore{}
}
