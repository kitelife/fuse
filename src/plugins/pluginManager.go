package plugins

import (
    "net/http"
)

type PluginInterface interface {
    Parse(*http.Request) (string, string)
}

var Plugins map[string]PluginInterface

func PluginRegister(thisPlugin PluginInterface) bool {
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

func Dispatch(pluginID string) (targetPlugin PluginInterface) {
    targetPlugin, ok := Plugins[pluginID]
    if ok == false {
        fmt.Println("不存在该插件！")
        return nil
    }
    return targetPlugin
}

func ListPluginID()(pluginIDList []string) {
    for pluginID, _ := range Plugins {
        pluginIDList = append(pluginIDList, pluginID)
    }
    return pluginIDList
}

func init() {
    Plugins = make(map[string]PluginInterface)
}
