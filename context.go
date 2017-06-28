package goplugin

import (
	"errors"
	"fmt"
	"sync"

	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/mq"
	"github.com/qxnw/lib4go/transform"
)

var contextPool *sync.Pool

func init() {
	contextPool = &sync.Pool{
		New: func() interface{} {
			return &PluginContext{}
		},
	}
}

type PluginContext struct {
	service     string
	ctx         Context
	Input       transform.ITransformGetter
	Params      transform.ITransformGetter
	Body        string
	Args        map[string]string
	GetVarValue func(c string, n string) (string, error)
	RPC         RPCInvoker
	producer    mq.MQProducer
	*logger.Logger
}

func (w *PluginContext) CheckMustFields(names ...string) error {
	for _, v := range names {
		if _, err := w.Input.Get(v); err != nil {
			err := fmt.Errorf("输入参数:%s不能为空", v)
			return err
		}
	}
	return nil
}

//CheckMapMustFields 检查map中的参数是否为空
func (w *PluginContext) CheckMapMustFields(input map[string]interface{}, names ...string) error {
	for _, v := range names {
		if input[v] == nil {
			err := fmt.Errorf("输入参数:%s不能为空", v)
			return err
		}
	}
	return nil
}

func GetContext(ctx Context, invoker RPCInvoker) (wx *PluginContext, err error) {
	wx = contextPool.Get().(*PluginContext)
	wx.ctx = ctx
	wx.producer = nil
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
	wx.RPC = invoker
	return
}

func (w *PluginContext) getLogger() (*logger.Logger, error) {
	if session, ok := w.ctx.GetExt()["hydra_sid"]; ok {
		return logger.GetSession("wx_base_core", session.(string)), nil
	}
	return nil, fmt.Errorf("输入的context里没有包含hydra_sid(%v)", w.ctx.GetExt())
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

func (w *PluginContext) Close() {
	if w.producer != nil {
		w.producer.Close()
	}
	contextPool.Put(w)
}
