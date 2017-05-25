package goplugin

type Context interface {
	GetInput() interface{}
	GetArgs() interface{}
	GetBody() interface{}
	GetParams() interface{}
	GetJson() string
	GetExt() map[string]interface{}
}
type RPCInvoker interface {
	//Request 发送请求
	Request(service string, input map[string]string, failFast bool) (status int, result string, err error)
	//Query 发送请求
	Query(service string, input map[string]string, failFast bool) (status int, result string, err error)
	//Update 发送请求
	Update(service string, input map[string]string, failFast bool) (status int, err error)
	//Insert 发送请求
	Insert(service string, input map[string]string, failFast bool) (status int, err error)
	//Delete 发送请求
	Delete(service string, input map[string]string, failFast bool) (status int, err error)
}

type PluginWorker interface {
	GetServices() []string
	Handle(name string, mode string, service string, c Context, invoker RPCInvoker) (status int, result string, err error)
}

//Handler 处理程序接口
type Handler interface {
	Handle(service string, c Context, invoker RPCInvoker) (status int, result string, err error)
}

//ServiceHandlers 服务处理程序列表
var ServiceHandlers map[string]Handler

//Services 当前提供的服务列表
var Services []string

func init() {
	ServiceHandlers = make(map[string]Handler)
	Services = make([]string, 0, 16)

}

//Register 注册处理程序
func Register(name string, handler Handler) {
	if _, ok := ServiceHandlers[name]; ok {
		panic("Register called twice for adapter " + name)
	}
	ServiceHandlers[name] = handler
	Services = append(Services, name)
}
