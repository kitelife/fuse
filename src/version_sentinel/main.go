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
    "plugins"
)

type Response struct {
    Status string
    Msg    string
}

type Repos struct {
    ReposID int,
    ReposName string
    ReposRemote string
}

type Hook struct {
    HookID int
    ReposID int
    WhichBranch string
    TargetDir   string
}

type Branch2Dir map[string]string

// var hooks map[int]Hook = make(map[int]Hook)

var masterAbsPath string
var db *sql.DB

func initDB() (err error) {
    // 如果目标数据表还不存在则创建
    tableRepos = `CREATE TABLE IF NOT EXISTS repos (
        repos_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        repos_name TEXT NOT NULL,
        repos_remote TEXT NOT NULL,
    )`
    _, err := db.Exec(tableRepos)
    
    if err != nil {
        return err
    }
    
    tableHooks = `CREATE TABLE IF NOT EXISTS hooks (
        hook_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        repos_id INTEGER NOT NULL,
        which_branch TEXT NOT NULL DEFAULT "master",
        target_dir TEXT NOT NULL,
        FOREIGN KEY(repos_id) REFERENCES repos(repos_id)
    );`
    _, err = db.Exec(tableHooks)
    if err != nil {
        return err
    }
    
    tableStatusLog = `CREATE TABLE IF NOT EXISTS status_log (
        log_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        hook_id INTEGER NOT NULL,
        status TEXT NOT NULL,
        log_content TEXT NOT NULL,
        updated_time TIMESTAMP,
        FORIGEN KEY (hook_id) REFERENCES hooks(hook_id)
    )`
    _, err = db.Exec(tableStatusLog)
    if err != nil {
        return err
    }
    return nil
}

func queryDBForHookHandler()(map[int]Repos, map[int]Branch2Dir) {
    repos := make(map[index]Repos)
    reposBranch2Dir := make(map[int]Branch2Dir)
    // 尝试读取数据
    hooksDataSQL = "SELECT hook_id, repos_id, which_branch, target_dir FROM hooks"
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
    return repos, reposBranch2Dir
}

func queryDataForViewHome() {
    
}

func genResponseStr(status string, message string) []byte {
    resp := Response{
        Status: status,
        Msg:    message,
    }
    responseContent, _ := json.MarshalIndent(resp, "", "    ")
    return responseContent
}

func HookHandler(w http.ResponseWriter, req *http.Request, params martini.Params) {
    w.Header().Set("Content-Type", "application/json")
    
    repos, reposBranch2Dir := queryDBForHookHandler()
    
    reposID := params["repos_id"]
    // 如果用户指定了代码库的远程地址，则使用指定的
    targetRepos, ok := repos[reposID]
    if ok == false {
        log.Fatalln("不存在指定的代码库", reposID)
        w.Write(genResponseStr("Failed", "不存在指定的代码库！"))
        return
    }
    reposRemoteURL := targetRepos.repos_remote
    
    pluginID := params["plugin_id"]
    // 根据请求中指定的插件ID，加载对应的插件
    targetPlugin := plugins.Dispatch(pluginID, req)
    if targetPlugin == nil {
        log.Fatalln("不存在指定的插件", pluginID)
        w.Write(genResponseStr("Failed", "请求的URL错误！"))
        return
    }
    remoteURL, branchName, err := targetPlugin.parse()
    if err != nil {
        log.Fatalln("请求内容解析出错！")
        w.Write(genResponseStr("Failed", "请求体不合法!"))
        return
    }
    if reposRemoteURL == nil || reposRemoteURL == "" {
        reposRemoteURL = remoteURL
    }
    
    thatBranch2Dir, _ := reposBranch2Dir[reposID]
    targetDir, ok := thatBranch2Dir[branchName]
    if ok == false {
        log.Fatalln("未针对该分支配置对应的Hook！", reposRemoteURL, branchName)
        w.Write(genResponseStr("Failed", "未针对该分支配置对应的Hook！"))
        return
    }

    isNew := false
    absTargetPath, _ := filepath.Abs(targetDir)
    if _, e := os.Stat(absTargetPath); os.IsNotExist(e) {
        os.Mkdir(absTargetDir, 0666)
        isNew = true
    }
    // 获取当前目录，用于切换回来
    pwd, _ := os.Getwd()
    // 切换当前目录到对应分支的代码目录
    err = os.Chdir(absTargetPath)
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

    err := initDB()
    if err != nil {
        fmt.Println("数据库操作失败！", err.Error())
        return
    }

    m := martini.Classic()

    m.Get("/", viewHome)
    m.Post("/webhook/(?P<plugin_id>[a-zA-Z]+)/(?P<repos_id>\d+)", hookEventHandler)
    m.Post("/new/repos", newRepos)
    m.Post("/new/hook", newHook)

    m.Run()
}
