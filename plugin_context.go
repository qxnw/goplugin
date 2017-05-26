package goplugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/lib4go/memcache"
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
	service      string
	ctx          Context
	Input        transform.ITransformGetter
	Params       transform.ITransformGetter
	Body         string
	db           *db.DB
	cache        *memcache.MemcacheClient
	Args         map[string]string
	func_var_get func(c string, n string) (string, error)
	RPC          RPCInvoker
	producer     mq.MQProducer
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

func GetContext(ctx Context, invoker RPCInvoker) (wx *PluginContext, err error) {
	wx = contextPool.Get().(*PluginContext)
	wx.ctx = ctx
	wx.db = nil
	wx.producer = nil

	defer func() {
		if err != nil {
			wx.Close()
		}
	}()
	wx.Input, err = wx.getGetParams(ctx.GetInput())
	if err != nil {
		return
	}
	wx.Params, err = wx.getGetParams(ctx.GetParams())
	if err != nil {
		return
	}
	wx.Body, err = wx.getGetBody(ctx.GetBody())
	if err != nil {
		return
	}
	wx.Args, err = wx.GetArgs(ctx.GetArgs())
	if err != nil {
		return
	}
	wx.func_var_get, err = wx.getVarParam(ctx.GetExt())
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

func (w *PluginContext) GetCache() (c *memcache.MemcacheClient, err error) {
	if w.cache != nil {
		return w.cache, nil
	}

	name, ok := w.Args["cache"]
	if !ok {
		return nil, fmt.Errorf("服务%s未配置cache参数(%v)", w.service, w.Args)
	}
	conf, err := w.func_var_get("cache", name)
	if err != nil {
		return nil, err
	}
	configMap, err := jsons.Unmarshal([]byte(conf))
	if err != nil {
		return nil, err
	}
	server, ok := configMap["server"]
	if !ok {
		err = fmt.Errorf("cache[%s]配置文件错误，未包含server节点:%s", name, conf)
		return nil, err
	}
	c, err = memcache.New(strings.Split(server.(string), ";"))
	if err != nil {
		return nil, err
	}
	w.cache = c
	return
}

func (w *PluginContext) GetJsonFromCache(tpl []string, input map[string]interface{}) (cvalue string, err error) {
	db, err := w.GetDB()
	if err != nil {
		return
	}
	sql, key, expireAt, err := w.getSqlKeyExpire(tpl)
	if err != nil {
		return
	}
	client, err := w.GetCache()
	if err != nil {
		return
	}
	tf := transform.NewMaps(input)
	key = tf.Translate(key)
	cvalue, err = client.Get(key)
	if err != nil {
		return
	}
	if cvalue != "" {
		return
	}
	data, _, _, err := db.Query(sql, input)
	if err != nil {
		return
	}
	buffer, err := jsons.Marshal(&data)
	if err != nil {
		return
	}
	client.Set(key, string(buffer), expireAt)
	return
}

func (w *PluginContext) GetFirstMapFromCache(tpl []string, input map[string]interface{}) (data map[string]interface{}, err error) {
	result, err := w.GetMapFromCache(tpl, input)
	if err != nil {
		return
	}
	if len(result) > 0 {
		return result[0], nil
	}
	return nil, fmt.Errorf("返回的数据条数为0:(%s)", tpl)
}

func (w *PluginContext) getSqlKeyExpire(tpl []string) (sql string, key string, expireAt int, err error) {
	if len(tpl) < 3 {
		err = fmt.Errorf("输入的SQL模板错误，必须包含3个元素，SQL语句/缓存KEY/过期时间:%v", tpl)
		return
	}
	sql = tpl[0]
	key = tpl[1]
	expireAt, err = strconv.Atoi(tpl[2])
	if err != nil {
		err = fmt.Errorf("输入的SQL模板错误，过期时间必须为数字:%v", tpl)
		return
	}
	return
}

func (w *PluginContext) getSql(tpl []string) (sql string, err error) {
	if len(tpl) >= 1 {
		err = fmt.Errorf("输入的SQL模板错误，必须包含1个元素，SQL语句:%v", tpl)
		return
	}
	sql = tpl[0]
	return
}
func (w *PluginContext) GetMapFromCache(tpl []string, input map[string]interface{}) (data []map[string]interface{}, err error) {
	sql, key, expireAt, err := w.getSqlKeyExpire(tpl)
	if err != nil {
		return
	}

	client, err := w.GetCache()
	if err != nil {
		return
	}
	tf := transform.NewMaps(input)
	key = tf.Translate(key)
	dstr, err := client.Get(key)
	if err != nil {
		return
	}
	if dstr != "" {
		err = json.Unmarshal([]byte(dstr), &data)
		return
	}
	db, err := w.GetDB()
	if err != nil {
		return
	}
	data, _, _, err = db.Query(sql, input)
	if err != nil {
		return
	}
	cvalue, err := jsons.Marshal(data)
	if err != nil {
		return
	}
	client.Set(key, string(cvalue), expireAt)
	return
}
func (w *PluginContext) ScalarFromDb(tpl []string, input map[string]interface{}) (data interface{}, err error) {
	db, err := w.GetDB()
	if err != nil {
		return
	}
	sql, err := w.getSql(tpl)
	if err != nil {
		return
	}
	data, _, _, err = db.Scalar(sql, input)
	return
}
func (w *PluginContext) ExecuteToDb(tpl []string, input map[string]interface{}) (row int64, err error) {
	db, err := w.GetDB()
	if err != nil {
		return
	}
	sql, err := w.getSql(tpl)
	if err != nil {
		return
	}
	row, _, _, err = db.Execute(sql, input)
	return
}
func (w *PluginContext) GetDataFromDb(tpl []string, input map[string]interface{}) (data []map[string]interface{}, err error) {
	db, err := w.GetDB()
	if err != nil {
		return
	}
	sql, err := w.getSql(tpl)
	if err != nil {
		return
	}
	data, _, _, err = db.Query(sql, input)
	return
}

func (w *PluginContext) GetDB() (d *db.DB, err error) {
	if w.db != nil {
		return w.db, nil
	}
	name, ok := w.Args["db"]
	if !ok {
		return nil, fmt.Errorf("服务%s未配置db参数(%v)", w.service, w.Args)
	}
	conf, err := w.func_var_get("db", name)
	if err != nil {
		return nil, err
	}
	configMap, err := jsons.Unmarshal([]byte(conf))
	if err != nil {
		return nil, err
	}
	provider, ok := configMap["provider"]
	if !ok {
		return nil, fmt.Errorf("db配置文件错误，未包含provider节点:var/db/%s", name)
	}
	connString, ok := configMap["connString"]
	if !ok {
		return nil, fmt.Errorf("db配置文件错误，未包含connString节点:var/db/%s", name)
	}
	d, err = db.NewDB(provider.(string), connString.(string), 2)
	if err != nil {
		err = fmt.Errorf("创建DB失败:err:%v", err)
		return
	}
	w.db = d
	return
}

func (w *PluginContext) getLogger() (*logger.Logger, error) {
	if session, ok := w.ctx.GetExt()["hydra_sid"]; ok {
		return logger.GetSession("wx_base_core", session.(string)), nil
	}
	return nil, fmt.Errorf("输入的context里没有包含hydra_sid(%v)", w.ctx.GetExt())
}
func (w *PluginContext) getVarParam(ext map[string]interface{}) (func(c string, n string) (string, error), error) {
	funcVar := ext["__func_var_get_"]
	if funcVar == nil {
		return nil, errors.New("未找到__func_var_get_")
	}
	if f, ok := funcVar.(func(c string, n string) (string, error)); ok {
		return f, nil
	}
	return nil, errors.New("未找到__func_var_get_传入类型错误")
}
func (w *PluginContext) GetArgs(args interface{}) (params map[string]string, err error) {
	params, ok := args.(map[string]string)
	if !ok {
		err = fmt.Errorf("未设置Args参数")
		return
	}
	return
}
func (w *PluginContext) getGetBody(body interface{}) (t string, err error) {
	if body == nil {
		return "", errors.New("body 数据为空")
	}
	t, ok := body.(string)
	if !ok {
		return "", errors.New("body 不是字符串数据")
	}
	return
}
func (w *PluginContext) getGetParams(input interface{}) (t transform.ITransformGetter, err error) {
	if input == nil {
		err = fmt.Errorf("输入参数为空:%v", input)
		return nil, err
	}
	t, ok := input.(transform.ITransformGetter)
	if !ok {
		return t, fmt.Errorf("输入参数为空:input（%v）不是transform.ITransformGetter类型", input)
	}
	return t, nil
}

//GetMQProducer 获取GetMQProducer
func (w *PluginContext) GetMQProducer() (p mq.MQProducer, err error) {
	if w.producer != nil {
		return w.producer, nil
	}
	name, ok := w.Args["mq"]
	if !ok {
		return nil, fmt.Errorf("服务%s未配置mq参数(%v)", w.service, w.Args)
	}
	conf, err := w.func_var_get("mq", name)
	if err != nil {
		return nil, err
	}
	configMap, err := jsons.Unmarshal([]byte(conf))
	if err != nil {
		return nil, err
	}
	address, ok := configMap["address"]
	if !ok {
		return nil, fmt.Errorf("mq配置文件错误，未包含provider节点:var/mq/%s", name)
	}
	p, err = mq.NewMQProducer(address.(string), mq.WithLogger(w.Logger))
	if err != nil {
		err = fmt.Errorf("创建mq失败:err:%v", err)
		return
	}
	err = p.Connect()
	if err != nil {
		err = fmt.Errorf("无法连接到MQ服务器:err:%v", err)
		return
	}
	w.producer = p

	return
}

func (w *PluginContext) Close() {
	if w.producer != nil {
		w.producer.Close()
	}
	contextPool.Put(w)
}
