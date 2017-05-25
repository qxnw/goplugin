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
	Handle(name string, mode string, service string, c Context, invoker RPCInvoker) (status int, result string, params map[string]interface{}, err error)
}

//Handler 处理程序接口
type Handler interface {
	Handle(service string, c Context, invoker RPCInvoker) (status int, result string, err error)
}

type Registry struct {
	ServiceHandlers map[string]Handler
	Services        []string
}

//NewRegistry 构建插件的注册中心
func NewRegistry() *Registry {
	r := &Registry{}
	r.ServiceHandlers = make(map[string]Handler)
	r.Services = make([]string, 0, 16)
	return r
}

//Register 注册处理程序
func (r *Registry) Register(name string, handler Handler) {
	if _, ok := r.ServiceHandlers[name]; ok {
		panic("Register called twice for adapter " + name)
	}
	r.ServiceHandlers[name] = handler
	r.Services = append(r.Services, name)
}
