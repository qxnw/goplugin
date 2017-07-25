package goplugin

import (
	"errors"
	"net/http"
)

//GetHttpRequest 获和取http.request对象
func (w *PluginContext) GetHttpRequest() (request *http.Request, err error) {

	r := w.ctx.GetExt()["__func_http_request_"]
	if r == nil {
		return nil, errors.New("未找到__func_http_request_")
	}
	if f, ok := r.(*http.Request); ok {
		return f, nil
	}
	return nil, errors.New("未找到__func_http_request_传入类型错误")
}
