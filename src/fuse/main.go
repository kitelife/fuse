package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "html/template"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"

    "github.com/go-martini/martini"
    _ "github.com/mattn/go-sqlite3"
    
    _ "plugins"
    "plugin_manager"
)

type ResponseStruct struct {
    Status string
    Msg    string
}

type ReposStruct struct {
    ReposID     int
    ReposName   string
    ReposRemote string
    ReposType string
}

type HookStruct struct {
    HookID      int
    ReposID     int
    WhichBranch string
    TargetDir   string
    HookStatus string
    LogContent string
    UpdatedTime string
}


type DBRelatedDataStruct struct {
    ReposStruct
    Hooks []HookStruct
}

type HomePageDataStruct struct {
    PluginIDList []string
    ReposList map[int]string
    DBRelatedData []DBRelatedDataStruct
}

type Branch2DirMap map[string]string

var masterAbsPath string
var db *sql.DB

func initDB() (err error) {
    // 如果目标数据表还不存在则创建
    tableRepos := `CREATE TABLE IF NOT EXISTS repos (
        repos_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        repos_name TEXT NOT NULL,
        repos_remote TEXT NOT NULL DEFAULT "",
        repos_type TEXT NOT NULL
    )`
    _, err = db.Exec(tableRepos)

    if err != nil {
        return err
    }

    tableHooks := `CREATE TABLE IF NOT EXISTS hooks (
        hook_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        repos_id INTEGER NOT NULL,
        which_branch TEXT NOT NULL DEFAULT "master",
        target_dir TEXT NOT NULL,
        hook_status TEXT NOT NULL,
        log_content TEXT NOT NULL,
        updated_time TIMESTAMP,
        FOREIGN KEY(repos_id) REFERENCES repos(repos_id)
    );`
    _, err = db.Exec(tableHooks)
    if err != nil {
        return err
    }
    return nil
}

func queryDBForHookHandler() (map[int]ReposStruct, map[int]Branch2DirMap) {
    // 尝试读取数据
    reposDataSQL := "SELECT repos_id, repos_name, repos_remote repos_type FROM repos"
    reposRows, err := db.Query(reposDataSQL)
    if err != nil {
        return nil, nil
    }
    defer reposRows.Close()

    hooksDataSQL := "SELECT repos_id, which_branch, target_dir FROM hooks"
    hooksRows, err := db.Query(hooksDataSQL)
    if err != nil {
        return nil, nil
    }
    defer hooksRows.Close()

    reposAll := make(map[int]ReposStruct)
    reposBranch2Dir := make(map[int]Branch2DirMap)

    var reposID int
    
    var reposName string
    var reposRemote string
    var reposType string
    for reposRows.Next() {
        reposRows.Scan(&reposID, &reposName, &reposRemote, &reposType)
        reposAll[reposID] = ReposStruct{ReposID: reposID, ReposName: reposName, ReposRemote: reposRemote, ReposType: reposType}
    }
    var whichBranch string
    var targetDir string
    for hooksRows.Next() {
        hooksRows.Scan(&reposID, &whichBranch, &targetDir)
        _, ok := reposBranch2Dir[reposID]
        if ok == false {
            reposBranch2Dir[reposID] = map[string]string{whichBranch: targetDir}
        } else {
            reposBranch2Dir[reposID][whichBranch] = targetDir
        }
    }
    return reposAll, reposBranch2Dir
}

func queryDataForViewHome()(reposList map[int]string, dbRelatedData []DBRelatedDataStruct) {
    reposList = make(map[int]string)
    
    reposDataSQL := "SELECT repos_id, repos_name, repos_remote, repos_type FROM repos"
    reposRows, err := db.Query(reposDataSQL)
    if err != nil {
        fmt.Println("数据库查询出错！", err.Error())
        return
    }
    hooksDataSQL := "SELECT hook_id, repos_id, which_branch, target_dir, hook_status, log_content, updated_time FROM hooks"
    hooksRows, err := db.Query(hooksDataSQL)
    if err != nil {
        fmt.Println("数据库查询出错！", err.Error())
        return
    }
    
    var reposID int
    
    var hookID int
    var whichBranch string
    var targetDir string
    var hookStatus string
    var logContent string
    var updatedTime string
    
    var hooks map[int][]HookStruct = make(map[int][]HookStruct)
    for hooksRows.Next() {
        hooksRows.Scan(&hookID, &reposID, &whichBranch, &targetDir, &hookStatus, &logContent, &updatedTime)
        if _, ok := hooks[reposID]; ok == false {
            // 1024: 每个代码库最多能够1024个分支，也即1024个hook
            hooks[reposID] = make([]HookStruct, 0, 1024)
        }
        //dbRelatedData[reposID].Hooks = append(dbRelatedData[reposID].Hooks, HookStruct{hookID, reposID, whichBranch, targetDir, hookStatus, logContent, updatedTime})
        hooks[reposID] = append(hooks[reposID], HookStruct{hookID, reposID, whichBranch, targetDir, hookStatus, logContent, updatedTime})
    }
    
    var reposName string
    var reposRemote string
    var reposType string
    for reposRows.Next() {
        reposRows.Scan(&reposID, &reposName, &reposRemote, &reposType)
        dbRelatedData = append(dbRelatedData, DBRelatedDataStruct{ReposStruct{reposID, reposName, reposRemote, reposType}, hooks[reposID]})
        reposList[reposID] = reposName
    }
    
    return reposList, dbRelatedData
}

func genResponseStr(status string, message string) []byte {
    resp := ResponseStruct{
        Status: status,
        Msg:    message,
    }
    responseContent, _ := json.MarshalIndent(resp, "", "    ")
    return responseContent
}

func hookEventHandler(w http.ResponseWriter, req *http.Request, params martini.Params) {
    w.Header().Set("Content-Type", "application/json")

    repos, reposBranch2Dir := queryDBForHookHandler()

    reposID, err := strconv.Atoi(params["repos_id"])
    if err != nil {
        fmt.Println("请求URL错误！")
        w.Write(genResponseStr("Failed", "请求的URL错误！"))
        return
    }
    // 如果用户指定了代码库的远程地址，则使用指定的
    targetRepos, ok := repos[reposID]
    if ok == false {
        fmt.Println("不存在指定的代码库", reposID)
        w.Write(genResponseStr("Failed", "不存在指定的代码库！"))
        return
    }
    reposRemoteURL := targetRepos.ReposRemote

    pluginID := params["plugin_id"]
    // 先检测请求URL中的仓库类型与目标仓库配置的类型是否一致
    if targetRepos.ReposType != pluginID {
        fmt.Println("仓库类型不匹配！")
        w.Write(genResponseStr("Failed", "请求的URL错误！"))
        return
    }
    // 根据请求中指定的插件ID，加载对应的插件
    targetPlugin := plugin_manager.Dispatch(pluginID)
    if targetPlugin == nil {
        fmt.Println("不存在指定的插件", pluginID)
        w.Write(genResponseStr("Failed", "请求的URL错误！"))
        return
    }
    remoteURL, branchName := targetPlugin.Parse(req)
    if reposRemoteURL == "" {
        reposRemoteURL = remoteURL
    }

    thatBranch2Dir, ok := reposBranch2Dir[reposID]
    if ok == false {
        fmt.Println("不存在指定代码库的Hook")
        return
    }
    targetDir, ok := thatBranch2Dir[branchName]
    if ok == false {
        fmt.Println("未针对该分支配置对应的Hook！", reposRemoteURL, branchName)
        w.Write(genResponseStr("Failed", "未针对该分支配置对应的Hook！"))
        return
    }

    isNew := false
    absTargetPath, _ := filepath.Abs(targetDir)
    if _, e := os.Stat(absTargetPath); os.IsNotExist(e) {
        os.Mkdir(absTargetPath, 0666)
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
            fmt.Println(err.Error())
            w.Write(genResponseStr("Failed", "克隆失败！"))
            return
        }
        fmt.Println(string(output))
    } else {
        // 先清除可能存在的本地变更
        cleanCMD := exec.Command("git", "checkout", "*")
        output, err := cleanCMD.Output()
        if err != nil {
            fmt.Println(err)
            w.Write(genResponseStr("Failed", "清除本地变更失败！"))
            return
        }
        fmt.Println(string(output))
    }

    changeBranchCMD := exec.Command("git", "checkout", branchName)
    output, err := changeBranchCMD.Output()
    if err != nil {
        fmt.Println(err)
        w.Write(genResponseStr("Failed", "工作目录切换到目标分支失败！"))
        return
    }
    fmt.Println(string(output))
    // 然后pull
    pullCmd := exec.Command("git", "pull", "-p")
    output, err = pullCmd.Output()
    if err != nil {
        fmt.Println(err)
        w.Write(genResponseStr("Failed", "Git Pull失败！"))
        return
    }
    fmt.Println(string(output))
    // 切换回原工作目录
    os.Chdir(pwd)
    w.Write(genResponseStr("success", "自动更新成功！"))
}

func viewHome(w http.ResponseWriter, req *http.Request) {
    pluginIDList := plugin_manager.ListPluginID()
    reposList, dbRelatedData := queryDataForViewHome()
    
    t, err := template.ParseFiles("./public/templates/index.html")
    if err != nil {
        fmt.Println(err)
        return
    }
    _ = t.Execute(w, HomePageDataStruct{PluginIDList: pluginIDList, ReposList: reposList, DBRelatedData: dbRelatedData})
    return
}

func newRepos() {

}

func newHook() {

}

func modifyRepos() {

}

func modifyHook() {

}

func deleteRepos() {

}

func deleteHook() {

}



func main() {
    var err error
    db, err = sql.Open("sqlite3", "./data.db")
    if err != nil {
        fmt.Println("数据库打开失败！", err.Error())
        return
    }
    defer db.Close()

    err = initDB()
    if err != nil {
        fmt.Println("数据库操作失败！", err.Error())
        return
    }

    m := martini.Classic()

    m.Get("/", viewHome)
    m.Post("/new/repos", newRepos)
    m.Post("/new/hook", newHook)
    m.Post("/modify/repos", modifyRepos)
    m.Post("/modify/hook", modifyHook)
    m.Post("/delete/repos", deleteRepos)
    m.Post("/delete/hook", deleteHook)
    
    m.Post("/webhook/(?P<plugin_id>[a-zA-Z]+)/(?P<repos_id>[0-9]+)", hookEventHandler)

    m.Run()
}
