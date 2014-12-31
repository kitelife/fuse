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
    "time"

    "github.com/go-martini/martini"
    _ "github.com/mattn/go-sqlite3"

    "plugin_manager"
    _ "plugins"
    "models"
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
var mh models.ModelHelper

func genResponseStr(status string, message string) []byte {
    resp := ResponseStruct{
        Status: status,
        Msg:    message,
    }
    responseContent, _ := json.MarshalIndent(resp, "", "    ")
    return responseContent
}

func checkPathExist(path string) bool {
    if _, e := os.Stat(absTargetPath); os.IsNotExist(e) {
        return false
    }
    return true
}

func hookEventHandler(w http.ResponseWriter, req *http.Request, params martini.Params) {
    w.Header().Set("Content-Type", "application/json")

    repos, reposBranch2Dir := mh.QueryDBForHookHandler()

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
    // 存入数据库时就已经是绝对路径了
    // absTargetPath, _ := filepath.Abs(targetDir)
    if checkPathExist(targetDir) == false {
        os.Mkdir(targetDir, 0666)
        isNew = true
    }
    // 获取当前目录，用于切换回来
    pwd, _ := os.Getwd()
    // 切换当前目录到对应分支的代码目录
    err = os.Chdir(targetDir)
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
    reposList, dbRelatedData := mh.QueryDBForViewHome()

    t, err := template.ParseFiles("./public/templates/index.html")
    if err != nil {
        fmt.Println(err)
        return
    }
    _ = t.Execute(w, HomePageDataStruct{PluginIDList: pluginIDList, ReposList: reposList, DBRelatedData: dbRelatedData})
    return
}

func newRepos(w http.ResponseWriter, req *http.Request, params martini.Params) {
    w.Header().Set("Content-Type", "application/json")

    reposName := params["repos_name"]

    if exist, _ := mh.CheckReposNameExists(db, reposName); exist == true {
        w.Write(genResponseStr("Failed", "该代码库已存在！"))
        return
    }

    reposType := params["repos_type"]
    if HasThisPlugin(reposType) == false {
        w.Write(genResponseStr("Failed", "不存在对应的代码库类型！"))
        return
    }

    reposRemote := params["repos_remote"]

    err := mh.StoreNewRepos(reposType, reposName, reposRemote)
    if err != nil {
        w.Write(genResponseStr("Failed", "新增代码库失败！"))
        return
    }
    w.Write(genResponseStr("success", "成功添加新代码库记录！"))
    return
}

func newHook(w http.ResponseWriter, params martini.Params) {
    reposID := strconv.Atoi(params["repos_id"])
    if exist, _ := mh.CheckReposIDExist(reposID); exist == false {
        w.Write(genResponseStr("Failed", "不存在指定的代码库！"))
        return
    }
    whichBranch := params["which_branch"]
    targetDir := params["target_dir"]
    // 存入数据库的是绝对路径
    targetDir, _ := filepath.Abs(targetDir)

    updatedTime := time.Now().String()
    if err := mh.StoreNewHook(reposID, whichBranch, targetDir, updatedTime); err != nil {
        w.Write(genResponseStr("Failed", "新增钩子失败！"))
        return
    }
    w.Write(genResponseStr("success", "成功添加新钩子！"))
    return
}

func deleteRepos(w http.ResponseWriter, params martini.Params) {
    reposID := strconv.Atoi(params["repos_id"])
    // 先检测是否还有hook关联到该代码库
    if exist, _ := mh.CheckReposHasHook(reposID); exist == true {
        w.Write(genResponseStr("Failed", "该代码库还关联有钩子！"))
        return
    }
    if err := mh.DeleteRepos(reposID); err != nil {
        w.Write(genResponseStr("Failed", "删除代码库记录失败！"))
        return
    }
    w.Write(genResponseStr("success", "成功删除代码库记录"))
    return
}

func deleteHook() {
    hookID := strconv.Atoi(params["hook_id"])
    // 是否彻底删除（包括代码目录）？
    eraseAll := params['erase_all']
    targetDir, err := mh.GetHookTargetDir(hookID)
    if err != nil {
        w.Write(genResponseStr("Failed", err.Error()))
        return
    }
    err = os.RemoveAll(targetDir)
    if err != nil {
        fmt.Println("删除代码目录出错，", err.Error())
    }
    if err := mh.DeleteHook(hookID); err != nil {
        w.Write(genResponseStr("Failed", "删除钩子失败！"))
        return
    }
    w.Write(genResponseStr("success", "成功删除钩子"))
    return
}

func main() {
    var err error
    db, err = sql.Open("sqlite3", "./data.db")
    if err != nil {
        fmt.Println("数据库打开失败！", err.Error())
        return
    }
    defer db.Close()

    mh = models.ModelHelper{Db: db}
    err = mh.initDB()
    if err != nil {
        fmt.Println("数据库操作失败！", err.Error())
        return
    }

    m := martini.Classic()

    m.Get("/", viewHome)
    m.Post("/new/repos", newRepos)
    m.Post("/new/hook", newHook)
    m.Post("/delete/repos", deleteRepos)
    m.Post("/delete/hook", deleteHook)

    m.Post("/webhook/(?P<plugin_id>[a-zA-Z]+)/(?P<repos_id>[0-9]+)", hookEventHandler)

    m.Run()
}
