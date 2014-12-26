package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/go-martini/martini"
    _ "github.com/mattn/go-sqlite3"
)

type ReposInfo struct {
    Name        string
    Url         string
    Description string
    Homepage    string
}

type CommitAuthorInfo struct {
    Name  string
    Email string
}

type CommitInfo struct {
    Id        string
    Message   string
    Timestamp string
    Url       string
    Author    CommitAuthorInfo
}

type PushRequestBody struct {
    Before     string
    After      string
    Ref        string
    User_id    int
    User_name  string
    Project_id int
    Repository ReposInfo
    Commits    []CommitInfo
}

type Response struct {
    Status string
    Msg    string
}

type Hook struct {
    TargetDir   string
    ReposName   string
    ReposRemote string
    WhichBranch string
}

type Branch2Dir map[string]string

var hooks map[string]Branch2Dir = make(map[string]Branch2Dir)
var repos map[string]string = make(map[string]string)
var masterAbsPath string
var db *sql.DB

func initFromDB() (err error) {
    // 如果目标数据表还不存在则创建
    tableHooks = `CREATE TABLE IF NOT EXISTS hooks (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        repos_name TEXT NOT NULL UNIQUE,
        repos_remote TEXT NOT NULL,
        which_branch TEXT NOT NULL DEFAULT "master",
        target_dir TEXT NOT NULL
    );`
    _, err := db.Exec(tableHooks)

    // 尝试读取数据
    hooksDataSQL = "SELECT repos_name, repos_remote, which_branch, target_dir FROM hooks"
    rows, err := db.Query(hooksDataSQL)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var targetDir string
    var reposName string
    var reposRemote string
    var whichBranch string
    for rows.Next() {
        rows.Scan(&targetDir, &reposName, &reposRemote, &whichBranch)
        hooks[&reposRemote] = map[string]string{&whichBranch: &targetDir}
        repos[&reposName] = &reposRemote
    }
    return nil
}

func genResponseStr(status string, message string) []byte {
    resp := Response{
        Status: status,
        Msg:    message,
    }
    responseContent, _ := json.MarshalIndent(resp, "", "    ")
    return responseContent
}

func HookHandler(w http.ResponseWriter, req *http.Request) {

    w.Header().Set("Content-Type", "application/json")

    // POST请求的处理
    var prb PushRequestBody
    eventDecoder := json.NewDecoder(req.Body)
    err := eventDecoder.Decode(&prb)
    if err != nil {
        log.Fatalln("请求内容非JSON格式：", err)
        w.Write(genResponseStr("Failed", "请求内容非JSON格式！"))
        return
    }

    // 记录日志
    reqBodyStr, _ := json.MarshalIndent(prb, "", "    ")
    // log.Println(string(reqBodyStr))

    reposRemoteURL := prb.Repository.Url
    branchParts := strings.Split(prb.Ref, "/")
    branchPartsLength = len(branchParts)
    if branchPartsLength == 0 {
        log.Fatalln("请求内容中分支不正确！", prb.Ref)
        w.Write(genResponseStr("Failed", "请求内容不合法！"))
        return
    }

    branchName := branchParts[branchPartsLength-1]
    thatBranch2Dir, ok := hooks[reposRemoteURL]
    if ok == false {
        log.Fatalln("未配置对应的Hook！", reposRemoteURL)
        w.Write(genResponseStr("Failed", "未配置对应的Hook！"))
        return
    }

    targetDir, ok := thatBranch2Dir[branchName]
    if ok == false {
        log.Fatalln("未针对该分支配置对应的Hook！", branchName)
        w.Write(genResponseStr("Failed", "未针对该分支配置对应的Hook！"))
        return
    }

    isNew := false
    absTargetDir, _ := filepath.Abs(targetDir)
    if _, e := os.Stat(absTargetDir); os.IsNotExist(e) {
        os.Mkdir(absTargetDir, 0666)
        isNew = true
    }
    // 获取当前目录，用于切换回来
    pwd, _ := os.Getwd()
    // 切换当前目录到master分支的代码目录
    err = os.Chdir(masterAbsPath)
    if err != nil {
        w.Write(genResponseStr("Error", "系统错误！"))
        return
    }

    if isNew {
        cloneCMD := exec.Command("git", "clone", reposRemoteURL, ".")
        output, err := cloneCMD.Output()
        if err != nil {
            log.Println(err.Error())
            w.Write(genResponseStr("Failed", "克隆失败！"))
            return
        }
        log.Println(string(output))
    } else {
        // 先清除可能存在的本地变更
        cleanCMD := exec.Command("git", "checkout", "*")
        output, err := cleanCMD.Output()
        if err != nil {
            log.Println(err)
            w.Write(genResponseStr("Failed", "清除本地变更失败！"))
            return
        }
        log.Println(string(output))
    }

    changeBranchCMD := exec.Command("git", "checkout", branchName)
    output, err = changeBranchCMD.Output()
    if err != nil {
        log.Println(err)
        w.Write(genResponseStr("Failed", "工作目录切换到目标分支失败！"))
        return
    }
    log.Println(string(output))
    // 然后pull
    pullCmd := exec.Command("git", "pull", "-p")
    output, err = pullCmd.Output()
    if err != nil {
        log.Println(err)
        w.Write(genResponseStr("Failed", "Git Pull失败！"))
        return
    }
    log.Println(string(output))
    // 切换回原工作目录
    os.Chdir(pwd)
    w.Write(genResponseStr("success", "自动更新成功！"))
}

func viewHome() string {

}

func newRepos() string {

}

func newHook() string {

}

func main() {
    db, err := sql.Open("sqlite3", "./data.db")
    if err != nil {
        fmt.Println("数据库打开失败！", err.Error())
        return
    }
    defer db.Close()

    err := initFromDB()
    if err != nil {
        fmt.Println("数据库操作失败！", err.Error())
        return
    }

    m := martini.Classic()

    m.Get("/", viewHome)
    m.Post("/webhook", hookEventHandler)
    m.Post("/new/repos", newRepos)
    m.Post("/new/hook", newHook)

    m.Run()
}
