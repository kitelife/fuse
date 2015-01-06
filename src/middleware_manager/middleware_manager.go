package middleware_manager

import (
    "models"
)

type MiddlewareInterface interface {
    Run(models.ChanElementStruct) (bool)
}

var Middlewares = map[string]MiddlewareInterface = make(map[string]MiddlewareInterface)

func MiddlewareRegister(id string, thisMiddleware MiddlewareInterface) bool {
    if _, ok := Middlewares[id]; ok == true {
        return false
    }
    Middlewares[id] = thisMiddleware
    return true
}

func Run(middlewareToRun []string, oneEvent models.ChanElementStruct) bool {
    for _, middlewareID := range middlewareToRun {
        // 返回false，表示执行失败，则不再执行接下来的中间件
        if Middlewares[middlewareID].Run(oneEvent) == false {
            return false
        }
    }
    return true
}

func init() {

}
