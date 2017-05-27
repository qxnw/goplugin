package goplugin

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/mq"
)

//SendMQMessage 发送MQ消息
func (w *PluginContext) SendMQMessage(queueName string, msg string, timeoutName string) error {
	if w.Args[queueName] == "" || w.Args[timeoutName] == "" {
		return fmt.Errorf("输入参数为空:%s或%s", queueName, timeoutName)
	}
	timeout, err := strconv.Atoi(w.Args[timeoutName])
	if err != nil {
		return fmt.Errorf("转换为数字失败.%s:%s", timeoutName, w.Args[timeoutName])
	}
	mqProducer, err := w.GetMQProducer()
	if err != nil {
		return fmt.Errorf("初始化MQ对象失败(err:%v)", err)
	}
	err = mqProducer.Send(w.Args[queueName], msg, time.Duration(timeout))
	if err != nil {
		return fmt.Errorf("发送MQ失败.队列名称:%s,消息内容:%s", w.Args[queueName], msg)
	}
	return nil

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
