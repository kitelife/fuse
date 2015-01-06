package utils

import (
    "fmt"
    "os"
)

func CheckPathExist(path string) bool {
    if _, e := os.Stat(path); os.IsNotExist(e) {
        return false
    }
    return true
}

func ChangeDir(targetDir string) {
    if err := os.Chdir(targetDir); err != nil {
        fmt.Println(err.Error())
    }
}
