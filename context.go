package goplugin

import (
	"errors"
	"fmt"
	"sync"

	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/memcache"
	"github.com/qxnw/lib4go/transform"
	"github.com/qxnw/lib4go/utility"
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
	logger.ILogger
}

//CheckInput 检查输入参数
func (w *PluginContext) CheckInput(names ...string) error {
	for _, v := range names {
		if r, err := w.Input.Get(v); err != nil || r == "" {
			err := fmt.Errorf("输入参数:%s不能为空", v)
			return err
		}
	}
	return nil
}

//CheckArgs 检查args参数
func (w *PluginContext) CheckArgs(names ...string) error {
	for _, v := range names {
		if r, ok := w.Args[v]; !ok || r == "" {
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
	wx.rpc = rpc
	wx.Input = ctx.GetInput()
	wx.Params = ctx.GetParams()
	wx.Args = ctx.GetArgs()
	wx.Body, _ = ctx.GetBody()
	wx.GetVarValue, err = wx.getVarParam()
	if err != nil {
		return
	}

	if _, ok := wx.ctx.GetExt()["__test__"]; ok {
		wx.ILogger = &tLogger{}
		return
	}
	wx.ILogger, err = wx.getLogger()
	if err != nil {
		return
	}
	return
}

func (w *PluginContext) getLogger() (*logger.Logger, error) {
	if session, ok := w.ctx.GetExt()["hydra_sid"]; ok {
		return logger.GetSession("wx_base_core", session.(string)), nil
	}
	return nil, fmt.Errorf("输入的context里没有包含hydra_sid(%v)", w.ctx.GetExt())
}

//ResetByBody 根据编码重置input参数
func (w *PluginContext) ResetByBody(encoding ...string) error {
	body, err := w.ctx.GetBody(encoding...)
	if err != nil {
		return err
	}
	qString, err := utility.GetMapWithQuery(body)
	if err != nil {
		return err
	}
	for k, v := range qString {
		w.Input.Set(k, v)
	}
	return nil
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

//GetString 从input获取字符串数据
func (w *PluginContext) GetString(name string) string {
	t, _ := w.Input.Get(name)
	return t
}

//GetSessionID session id
func (w *PluginContext) GetSessionID() string {
	return w.ILogger.GetSessionID()
}

//GetVarParam 获取var参数值，需提供在ext中提供__func_var_get_
func (w *PluginContext) GetVarParam(tpName string, name string) (string, error) {
	func_var := w.ctx.GetExt()["__func_var_get_"]
	if func_var == nil {
		return "", errors.New("未找到__func_var_get_")
	}
	if f, ok := func_var.(func(c string, n string) (string, error)); ok {
		s, err := f(tpName, name)
		if err != nil {
			err = fmt.Errorf("无法通过获取到参数/@domain/var/%s/%s的值", tpName, name)
			return "", err
		}
		return s, nil
	}
	return "", errors.New("未找到__func_var_get_传入类型错误")
}

//GetArgByName 获取arg的参数
func (w *PluginContext) GetArgByName(name string) (string, error) {
	argsMap := w.ctx.GetArgs()
	db, ok := argsMap[name]
	if db == "" || !ok {
		return "", fmt.Errorf("args配置错误，缺少:%s参数:%v", name, w.ctx.GetArgs())
	}
	return db, nil
}

//GetVarParamByArgsName 根据args参数名获取var参数的值
func (w *PluginContext) GetVarParamByArgsName(tpName string, argsName string) (string, error) {
	name, err := w.GetArgByName(argsName)
	if err != nil {
		return "", err
	}
	return w.GetVarParam(tpName, name)
}
