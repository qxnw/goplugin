package goplugin

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qxnw/lib4go/concurrent/cmap"
	"github.com/qxnw/lib4go/jsons"
	"github.com/qxnw/lib4go/mq"
)

//ContextMQ MQ操作实例
type ContextMQ struct {
	ctx *PluginContext
}

//Reset 重置context
func (cmq *ContextMQ) Reset(ctx *PluginContext) {
	cmq.ctx = ctx
}

//Send 发送MQ消息
func (cmq *ContextMQ) Send(queue string, msg string, timeout int) error {
	mqProducer, err := cmq.GetProducer()
	if err != nil {
		return fmt.Errorf("初始化MQ对象失败(err:%v)", err)
	}
	err = mqProducer.Send(queue, msg, time.Duration(timeout)*time.Second)
	if err != nil {
		return fmt.Errorf("发送MQ消息失败.队列名称:%s,消息内容:%s", queue, msg)
	}
	return nil

}

//SendBySetting 发送MQ消息
func (cmq *ContextMQ) SendBySetting(queueName string, msg string, timeoutName string) error {
	if cmq.ctx.Args[queueName] == "" || cmq.ctx.Args[timeoutName] == "" {
		return fmt.Errorf("args配置缺少MQ参数:%s，%s", queueName, timeoutName)
	}
	timeout, err := strconv.Atoi(cmq.ctx.Args[timeoutName])
	if err != nil {
		return fmt.Errorf("args配置的MQ超时时长必须为数字.名称：%s，值:%s", timeoutName, cmq.ctx.Args[timeoutName])
	}
	return cmq.Send(cmq.ctx.Args[queueName], msg, timeout)

}

//GetProducer GetProducer
func (cmq *ContextMQ) GetProducer() (p mq.MQProducer, err error) {
	name, ok := cmq.ctx.Args["mq"]
	if !ok {
		return nil, fmt.Errorf("args中未配置参数:mq,%v", cmq.ctx.Args)
	}
	_, imq, err := mqCache.SetIfAbsentCb(name, func(input ...interface{}) (d interface{}, err error) {
		name := input[0].(string)
		conf, err := cmq.ctx.GetVarValue("mq", name)
		if err != nil {
			err = fmt.Errorf("未找到mq配置参数:var/mq/%s,err:%v", name, err)
			return nil, err
		}
		configMap, err := jsons.Unmarshal([]byte(conf))
		if err != nil {
			err = fmt.Errorf("var/mq/%s配置参数的值不是有效的json err:%v", name, err)
			return nil, err
		}
		address, ok := configMap["address"]
		if !ok {
			return nil, fmt.Errorf("mq配置文件错误，未包含address节点:var/mq/%s", name)
		}
		p, err := mq.NewMQProducer(address.(string), mq.WithLogger(cmq.ctx.ILogger))
		if err != nil {
			err = fmt.Errorf("创建mq失败,%s:err:%v", address, err)
			return
		}
		err = p.Connect()
		if err != nil {
			err = fmt.Errorf("无法连接到MQ服务器:%v,err:%v", address, err)
			return
		}
		return p, err
	}, name)
	if err != nil {
		return
	}
	p = imq.(mq.MQProducer)
	return

}

var mqCache cmap.ConcurrentMap

func init() {
	mqCache = cmap.New(2)
}
