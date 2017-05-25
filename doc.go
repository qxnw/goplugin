package goplugin
/*
//编译指令
//go build -buildmode=plugin
import (
	"fmt"

	"github.com/qxnw/hydra_plugin/plugins"
	"github.com/qxnw/wx_base_core_api/handlers"
)

type wxBaseCore struct {
}

//GetServices 获取当前插件提供的所有服务
func (p *wxBaseCore) GetServices() []string {
	return handlers.Services
}

//Handle 业务处理
func (p *wxBaseCore) Handle(name string, mode string, service string, c plugins.Context, invoker plugins.RPCInvoker) (status int, result string, err error) {
	if h, ok := handlers.ServiceHandlers[service]; ok {
		status, result, err = h.Handle(service, c, invoker)
		if err != nil {
			return status, result, fmt.Errorf("wx_base_core_api:%v", err)
		}
		return
	}
	return 404, "", fmt.Errorf("wx_base_core_api 未找到服务:%s", service)
}

//GetWorker 获取当前worker
func GetWorker() plugins.PluginWorker {
	return &wxBaseCore{}
}
*/