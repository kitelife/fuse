package plugins

type Plugin interface {
    IsMe(req *http.Request) bool
    Parse()(string, string)
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

func Dispatch(req *http.Request)(targetPlugin Plugin) {
    for _, p := range Plugins {
        if p.IsMe(req) == true {
            targetPlugin = p
            break
        }
    }
    return targetPlugin
}

func init() {
    Plugins = make(map[string]Plugin)
}