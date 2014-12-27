package plugins

type Plugin interface {
	Parse() (string, string)
}

var Plugins map[string]Plugin

func PluginRegister(thisPlugin Plugin) bool {
	_, ok := Plugins[id]
	if ok == true {
		return false
	}
	if thisPlugin.id == nil {
		return false
	}
	Plugins[thisPlugin.id] = thisPlugin
	return true
}

func Dispatch(pluginID string, req *http.Request) (targetPlugin Plugin) {
	targetPlugin, ok := Plugins[pluginID]
	if ok == false {
		fmt.Println("不存在该插件！")
		return nil
	}
	return targetPlugin
}

func init() {
	Plugins = make(map[string]Plugin)
}
