package goplugin

import (
	"fmt"

	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/mq"
)

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
