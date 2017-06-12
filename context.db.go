package goplugin

import (
	"fmt"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/db"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/transform"
)

func (w *PluginContext) ScalarFromDb(tpl []string, input map[string]interface{}) (data interface{}, err error) {
	db, err := w.GetDB()
	if err != nil {
		return
	}
	sql, _, err := w.getSql(tpl)
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
	sql, key, err := w.getSql(tpl)
	if err != nil {
		return
	}
	row, _, _, err = db.Execute(sql, input)
	if err != nil {
		return
	}
	if key != "" {
		c, err := w.GetCache()
		if err != nil {
			w.Error("清除缓存失败:", err)
			return row, nil
		}
		tf := transform.NewMaps(input)
		err = c.Delete(tf.Translate(key))
		if err != nil {
			w.Error("清除缓存失败:", err)
		}
	}
	return
}
func (w *PluginContext) GetDataFromDb(tpl []string, input map[string]interface{}) (data []db.QueryRow, err error) {
	db, err := w.GetDB()
	if err != nil {
		return
	}
	sql, _, err := w.getSql(tpl)
	if err != nil {
		return
	}
	data, _, _, err = db.Query(sql, input)
	return
}

func (w *PluginContext) GetDB() (d *db.DB, err error) {
	name, ok := w.Args["db"]
	if !ok {
		return nil, fmt.Errorf("服务%s未配置db参数(%v)", w.service, w.Args)
	}
	_, dbc, err := dbCache.SetIfAbsentCb(name, func(input ...interface{}) (d interface{}, err error) {
		name := input[0].(string)
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
		return
	}, name)
	if err != nil {
		return
	}
	d = dbc.(*db.DB)
	return
}

var dbCache cmap.ConcurrentMap

func init() {
	dbCache = cmap.New()
}
