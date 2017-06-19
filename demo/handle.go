package main

import (
	"fmt"

	"strings"

	"github.com/qxnw/goplugin"
	"qxnw.com/cbs_core/lib/utility"
)

type sendSms struct {
	mustFields []string
}

func NewSendSms() *sendSms {
	return &sendSms{
		mustFields: []string{},
	}
}
func (n *sendSms) initParams(ctx goplugin.Context, invoker goplugin.RPCInvoker) (context *goplugin.PluginContext, status int, err error) {
	context, err = goplugin.GetContext(ctx, invoker)
	if err != nil {
		status = utility.SERVER_ERROR
		return
	}
	return
}

func (n *sendSms) Handle(service string, ctx goplugin.Context, invoker goplugin.RPCInvoker) (status int, result interface{}, err error) {
	context, status, err := n.initParams(ctx, invoker)
	if err != nil {
		return
	}

	data, err := context.GetDataFromDb(QUERY_SQL, make(map[string]interface{}))
	if err != nil {
		return
	}
	for _, v := range data {
		mobile := v.GetString("mobile")
		input := map[string]string{
			"mobile": mobile,
			"data":   "3333",
		}
		_, rs, err := invoker.Request("/sms/ytx/send", input, true)
		if err != nil {
			return 500, nil, err
		}
		input2 := map[string]interface{}{
			"mobile": mobile,
		}
		r, err := context.ExecuteToDb(UPDATE_SQL, input2)
		if err != nil {
			return 500, nil, err
		}
		if r != 1 {
			return 500, nil, fmt.Errorf("发送失败:%v", r)
		}
		if strings.Contains(rs, "000000") {
			context.Logger.Infof("发送成功:%s", mobile)
		} else {
			context.Logger.Errorf("发送失败:%s-%s", mobile, rs)
		}

	}

	return utility.OK, data, nil
}

var QUERY_SQL = []string{
	"SELECT mobile FROM SYS_MOBILE WHERE STATUS=1 and rownum<=10",
	"---",
	"0"}
var UPDATE_SQL = []string{
	"update SYS_MOBILE set status=0 WHERE mobile=@mobile",
	"",
	""}
