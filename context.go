package goplugin

import (
	"errors"
	"fmt"
	"sync"

	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/memcache"
	"github.com/qxnw/lib4go/transform"
)

var contextPool *sync.Pool

func init() {
	contextPool = &sync.Pool{
		New: func() interface{} {
			return &PluginContext{
				DB:    &ContextDB{},
				Cache: &ContextCache{},
				MQ:    &ContextMQ{},
				RPC:   &ContextRPC{},
			}
		},
	}
}

//PluginContext 插件上下文件
type PluginContext struct {
	ctx         Context
	Input       transform.ITransformGetter
	Params      transform.ITransformGetter
	Body        string
	Args        map[string]string
	GetVarValue func(c string, n string) (string, error)
	rpc         RPCInvoker
	DB          *ContextDB
	Cache       *ContextCache
	MQ          *ContextMQ
	RPC         *ContextRPC
	*logger.Logger
}

//CheckInput 检查输入参数
func (w *PluginContext) CheckInput(names ...string) error {
	for _, v := range names {
		if _, err := w.Input.Get(v); err != nil {
			err := fmt.Errorf("输入参数:%s不能为空", v)
			return err
		}
	}
	return nil
}

//CheckArgs 检查args参数
func (w *PluginContext) CheckArgs(names ...string) error {
	for _, v := range names {
		if _, ok := w.Args[v]; !ok {
			err := fmt.Errorf("args配置中缺少参数:%s", v)
			return err
		}
	}
	return nil
}

//CheckMap 检查map中的参数是否为空
func (w *PluginContext) CheckMap(input map[string]interface{}, names ...string) error {
	for _, v := range names {
		if input[v] == nil {
			err := fmt.Errorf("输入参数:%s不能为空", v)
			return err
		}
	}
	return nil
}

//GetContext 根据输入的context创建插件的上下文对象
func GetContext(ctx Context, rpc RPCInvoker) (wx *PluginContext, err error) {
	wx = contextPool.Get().(*PluginContext)
	wx.ctx = ctx
	wx.Cache.Reset(wx)
	wx.DB.Reset(wx)
	wx.MQ.Reset(wx)
	wx.RPC.Reset(wx)
	defer func() {
		if err != nil {
			wx.Close()
		}
	}()
	wx.Input = ctx.GetInput()
	wx.Params = ctx.GetParams()
	wx.Args = ctx.GetArgs()
	wx.Body = ctx.GetBody()
	wx.GetVarValue, err = wx.getVarParam()
	if err != nil {
		return
	}
	wx.Logger, err = wx.getLogger()
	if err != nil {
		return
	}
	wx.rpc = rpc
	return
}

func (w *PluginContext) getLogger() (*logger.Logger, error) {
	if session, ok := w.ctx.GetExt()["hydra_sid"]; ok {
		return logger.GetSession("wx_base_core", session.(string)), nil
	}
	return nil, fmt.Errorf("输入的context里没有包含hydra_sid(%v)", w.ctx.GetExt())
}

//GetCache 获取缓存操作对象
func (w *PluginContext) GetCache() (c *memcache.MemcacheClient, err error) {
	return w.Cache.GetCache()
}

//GetDB 获取数据库操作实例
func (w *PluginContext) GetDB() (d *db.DB, err error) {
	return w.DB.GetDB()
}
func (w *PluginContext) getVarParam() (func(c string, n string) (string, error), error) {

	funcVar := w.ctx.GetExt()["__func_var_get_"]
	if funcVar == nil {
		return nil, errors.New("未找到__func_var_get_")
	}
	if f, ok := funcVar.(func(c string, n string) (string, error)); ok {
		return f, nil
	}
	return nil, errors.New("未找到__func_var_get_传入类型错误")
}

//Close 回收当前context对象
func (w *PluginContext) Close() {
	contextPool.Put(w)
}
