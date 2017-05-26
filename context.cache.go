package goplugin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/memcache"
	"github.com/qxnw/lib4go/transform"
)

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
