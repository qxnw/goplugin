package goplugin

type GoPluginError struct {
	error
	data interface{}
}
