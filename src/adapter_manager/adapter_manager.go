package adapter_manager

import (
    "fmt"
    "net/http"
)

type FilteredEventDataStruct struct {
    ReposRemoteURL string
    BranchName string
    LatestCommit string
}

type AdapterInterface interface {
    Parse(*http.Request) (FilteredEventDataStruct)
}

var Adapters map[string]AdapterInterface = make(map[string]AdapterInterface)

func AdapterRegister(id string, thisAdapter AdapterInterface) bool {
    if _, ok := Adapters[id]; ok == true {
        return false
    }
    Adapters[id] = thisAdapter
    return true
}

func Dispatch(adapterID string) (targetAdapter AdapterInterface) {
    var ok bool
    if targetAdapter, ok = Adapters[adapterID]; ok == false {
        fmt.Println("不存在该适配器！")
        return nil
    }
    return targetAdapter
}

func ListAdapterID()(adapterIDList []string) {
    for adapterID, _ := range Adapters {
        adapterIDList = append(adapterIDList, adapterID)
    }
    return adapterIDList
}

func HasThisAdapter(adapterID string) bool {
    if _, ok := Adapters[adapterID]; ok == false {
        return false
    }
    return true
}

func init() {

}
