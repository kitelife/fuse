package plugin_manager

import (
    "fmt"
    "net/http"
)

type PluginInterface interface {
    Parse(*http.Request) (string, string)
}

var Plugins map[string]PluginInterface = make(map[string]PluginInterface)

func PluginRegister(id string, thisPlugin PluginInterface) bool {
    if _, ok := Plugins[id]; ok == true {
        return false
    }
    Plugins[id] = thisPlugin
    return true
}

func Dispatch(pluginID string) (targetPlugin PluginInterface) {
    var ok bool
    if targetPlugin, ok = Plugins[pluginID]; ok == false {
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

func HasThisPlugin(pluginID string) bool {
    if _, ok := Plugins[pluginID]; ok == false {
        return false
    }
    return true
}

func init() {
    
}
