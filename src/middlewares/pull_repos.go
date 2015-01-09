package middlewares

import (
    "fmt"
    "os"
    "os/exec"

    "middleware_manager"
    "models"
    "utils"
)

type PullReposStruct struct {
    id string
}

func (pr PullReposStruct) Run(chanElement models.ChanElementStruct) bool {
    isNew := false
    // 存入数据库时就已经是绝对路径了
    // absTargetPath, _ := filepath.Abs(targetDir)
    if utils.CheckPathExist(chanElement.TargetDir) == false {
        // 创建新目录可能会失败
        if err := os.MkdirAll(chanElement.TargetDir, 0666); err != nil {
            // 记录状态用于页面展示
            // 三种状态：error、failure、success
            // error 表示系统错误
            // failure表示pull/clone出错
            chanElement.Mh.UpdateLogStatus(chanElement.HookID, "error", err.Error())
            fmt.Println(err.Error())
            return false
        }
        isNew = true
    }

    // 获取当前目录，用于切换回来
    pwd, _ := os.Getwd()
    // 切换当前目录到对应分支的代码目录
    // 对于目录的切换，在goroutine运行在多线程的情况下可能会有问题
    if err := os.Chdir(chanElement.TargetDir); err != nil {
        chanElement.Mh.UpdateLogStatus(chanElement.HookID, "error", err.Error())
        fmt.Println(err.Error())
        return false
    }

    // 确保函数执行结束后能切换回原工作目录
    defer utils.ChangeDir(pwd)

    if isNew {
        cloneCMD := exec.Command("git", "clone", chanElement.RemoteURL, ".")
        output, err := cloneCMD.Output()
        if err != nil {
            errMsg := fmt.Sprintf("%s; %s", string(output), err.Error())
            chanElement.Mh.UpdateLogStatus(chanElement.HookID, "failure", errMsg)
            fmt.Println(errMsg)
            return false
        }
    } else {
        // 先清除可能存在的本地变更
        cleanCMD := exec.Command("git", "checkout", "*")
        output, err := cleanCMD.Output()
        if err != nil {
            errMsg := fmt.Sprintf("%s; %s", string(output), err.Error())
            chanElement.Mh.UpdateLogStatus(chanElement.HookID, "failure", errMsg)
            fmt.Println(errMsg)
            return false
        }
    }

    changeBranchCMD := exec.Command("git", "checkout", chanElement.BranchName)
    output, err := changeBranchCMD.Output()
    if err != nil {
        errMsg := fmt.Sprintf("%s; %s", string(output), err.Error())
        chanElement.Mh.UpdateLogStatus(chanElement.HookID, "failure", errMsg)
        fmt.Println(errMsg)
        return false
    }

    if isNew == false {
        pullCMD := exec.Command("git", "pull", "-p")
        output, err = pullCMD.Output()
        if err != nil {
            errMsg := fmt.Sprintf("%s; %s", string(output), err.Error())
            chanElement.Mh.UpdateLogStatus(chanElement.HookID, "failure", errMsg)
            fmt.Println(errMsg)
            return false
        }

        resetHardCMD := exec.Command("git", "reset", "--hard", chanElement.LatestCommit)
        output, err = resetHardCMD.Output()
        if err != nil {
            errMsg := fmt.Sprintf("%s; %s", string(output), err.Error())
            chanElement.Mh.UpdateLogStatus(chanElement.HookID, "failure", errMsg)
            fmt.Println(errMsg)
            return false
        }

    }
    chanElement.Mh.UpdateLogStatus(chanElement.HookID, "success", "成功拉取代码库！")
    return true
}

func init() {
    pr := PullReposStruct{id: "pull_repos"}
    middleware_manager.MiddlewareRegister("pull_repos", pr)
}
