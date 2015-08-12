package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	_ "github.com/mattn/go-sqlite3"

	"adapter_manager"
	_ "adapters"
	"config"
	"middleware_manager"
	_ "middlewares"
	"models"
)

var runTimeBasePath string
var db *sql.DB
var mh models.ModelHelper
var conf config.ConfStruct

// 带缓冲的事件数据管道
var eventChannels models.ReposChanMap

// 用来传递信号告知goroutine退出
var signalChannels = make(map[int]chan int)

func genResponseStr(status string, message string) []byte {
	resp := models.ResponseStruct{
		Status: status,
		Msg:    message,
	}
	responseContent, _ := json.MarshalIndent(resp, "", "    ")
	return responseContent
}

func hookEventHandler(w http.ResponseWriter, req *http.Request, params martini.Params) {
	w.Header().Set("Content-Type", "application/json")

	repos, reposBranch2Hook := mh.QueryDBForHookHandler()

	reposID, err := strconv.Atoi(params["repos_id"])
	if err != nil {
		fmt.Println("请求URL错误！")
		w.Write(genResponseStr("failure", "请求的URL错误！"))
		return
	}
	// 如果用户指定了代码库的远程地址，则使用指定的
	targetRepos, ok := repos[reposID]
	if ok == false {
		fmt.Println("不存在指定的代码库", reposID)
		w.Write(genResponseStr("failure", "不存在指定的代码库！"))
		return
	}
	reposRemoteURL := targetRepos.ReposRemote

	adapterID := params["adapter_id"]
	// 先检测请求URL中的仓库类型与目标仓库配置的类型是否一致
	if targetRepos.ReposType != adapterID {
		fmt.Println("仓库类型不匹配！")
		w.Write(genResponseStr("failure", "请求的URL错误！"))
		return
	}
	// 根据请求中指定的插件ID，加载对应的插件
	targetAdapter := adapter_manager.Dispatch(adapterID)
	if targetAdapter == nil {
		fmt.Println("不存在指定的适配器", adapterID)
		w.Write(genResponseStr("failure", "请求的URL错误！"))
		return
	}
	filteredEventData, err := targetAdapter.Parse(req)
	if err != nil {
		fmt.Println(err.Error())
		w.Write(genResponseStr("failure", err.Error()))
		return
	}
	if reposRemoteURL == "" {
		reposRemoteURL = filteredEventData.ReposRemoteURL
	}
	branchName := filteredEventData.BranchName

	thatBranch2Hook, ok := reposBranch2Hook[reposID]
	if ok == false {
		fmt.Println("不存在指定代码库的Hook")
		return
	}
	targetHookList, ok := thatBranch2Hook[branchName]
	if ok == false {
		fmt.Println("未针对该分支配置对应的Hook！", reposRemoteURL, branchName)
		w.Write(genResponseStr("failure", "未针对该分支配置对应的Hook！"))
		return
	}

	for _, oneHook := range targetHookList {
		newEventData := models.ChanElementStruct{
			ReposID:      reposID,
			HookID:       oneHook.HookID,
			RemoteURL:    reposRemoteURL,
			BranchName:   branchName,
			LatestCommit: filteredEventData.LatestCommit,
			TargetDir:    oneHook.TargetDir,
			Mh:           mh,
		}
		// 将事件数据传入管道
		eventChannels[reposID] <- newEventData
	}

	w.Write(genResponseStr("success", "成功！"))
	return
}

func viewHome(w http.ResponseWriter, req *http.Request) {
	adapterIDList := adapter_manager.ListAdapterID()
	reposList, dbRelatedData := mh.QueryDBForViewHome()

	t, err := template.ParseFiles(filepath.Join(runTimeBasePath, "public/templates/index.html"))
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = t.Execute(w, models.HomePageDataStruct{AdapterIDList: adapterIDList, ReposList: reposList, DBRelatedData: dbRelatedData})
	return
}

func newRepos(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reposName := req.FormValue("repos_name")

	if exist, _ := mh.CheckReposNameExists(db, reposName); exist == true {
		w.Write(genResponseStr("failure", "该代码库已存在！"))
		return
	}

	reposType := req.FormValue("repos_type")

	if adapter_manager.HasThisAdapter(reposType) == false {
		w.Write(genResponseStr("failure", "不存在对应的代码库类型！"))
		return
	}

	reposRemote := req.FormValue("repos_remote")

	newReposID, err := mh.StoreNewRepos(reposType, reposName, reposRemote)
	if err != nil {
		w.Write(genResponseStr("failure", "新增代码库失败！"))
		return
	}
	// 开启对应的goroutine和channel
	eventChannels[newReposID] = make(chan models.ChanElementStruct, conf.Queue_length)
	signalChannels[newReposID] = make(chan int)
	go HookWorker(eventChannels[newReposID], signalChannels[newReposID])

	w.Write(genResponseStr("success", "成功添加新代码库记录！"))
	return
}

func newHook(w http.ResponseWriter, req *http.Request) {
	reposID, _ := strconv.Atoi(req.FormValue("repos_id"))
	if exist, _ := mh.CheckReposIDExist(reposID); exist == false {
		w.Write(genResponseStr("failure", "不存在指定的代码库！"))
		return
	}
	whichBranch := req.FormValue("which_branch")
	targetDir := req.FormValue("target_dir")
	// 存入数据库的是绝对路径
	targetDir, _ = filepath.Abs(targetDir)

	updatedTime := time.Now().UTC().Format("2006-01-02 15:04:05")
	if _, err := mh.StoreNewHook(reposID, whichBranch, targetDir, updatedTime); err != nil {
		fmt.Println(err)
		w.Write(genResponseStr("failure", "新增钩子失败！"))
		return
	}
	w.Write(genResponseStr("success", "成功添加新钩子！"))
	return
}

func deleteRepos(w http.ResponseWriter, req *http.Request) {
	reposID, _ := strconv.Atoi(req.FormValue("repos_id"))
	// 先检测是否还有hook关联到该代码库
	if exist, _ := mh.CheckReposHasHook(reposID); exist == true {
		w.Write(genResponseStr("failure", "该代码库还关联有分支hook！"))
		return
	}
	if err := mh.DeleteRepos(reposID); err != nil {
		w.Write(genResponseStr("failure", "删除代码库记录失败！"))
		return
	}

	// 关闭对应的goroutine和channel
	signalChannels[reposID] <- 0
	delete(eventChannels, reposID)
	delete(signalChannels, reposID)

	w.Write(genResponseStr("success", "成功删除代码库记录"))
	return
}

func deleteHook(w http.ResponseWriter, req *http.Request) {
	hookID, _ := strconv.Atoi(req.FormValue("hook_id"))
	// 是否彻底删除（包括代码目录）？
	eraseAll := req.FormValue("erase_all")
	targetDir, err := mh.GetHookTargetDir(hookID)
	if err != nil {
		w.Write(genResponseStr("failure", err.Error()))
		return
	}
	if eraseAll == "true" {
		if err = os.RemoveAll(targetDir); err != nil {
			fmt.Println("删除代码目录出错，", err.Error())
		}
	}

	if err = mh.DeleteHook(hookID); err != nil {
		w.Write(genResponseStr("failure", "删除钩子失败！"))
		return
	}

	w.Write(genResponseStr("success", "成功删除钩子"))
	return
}

func HookWorker(eventChan chan models.ChanElementStruct, signalChan chan int) {
	fmt.Println("New HookWorker goroutine is running!")
	for {
		select {
		case oneEvent := <-eventChan:
			fmt.Println("收到事件，", oneEvent)
			if middleware_manager.Run(conf.Middlewares, oneEvent) == false {
				fmt.Println(oneEvent, "事件处理失败！")
			}
		case <-signalChan:
			close(signalChan)
			close(eventChan)
			fmt.Println("Goroutine接收到退出信号！")
			return
		default:
			//
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func RunWorkers() {
	for reposID, _ := range eventChannels {
		signalChannels[reposID] = make(chan int)
		go HookWorker(eventChannels[reposID], signalChannels[reposID])
	}
}

func withAuthOrNot() (authHandler martini.Handler) {
	if conf.Auth.Use {
		return auth.Basic(conf.Auth.Username, conf.Auth.Password)
	}
	return
}

func main() {
	var err error

	runTimeBasePath, err = os.Getwd()
	if err != nil {
		fmt.Println("获取当前路径失败！", err.Error())
		return
	}

	db, err = sql.Open("sqlite3", filepath.Join(runTimeBasePath, "data.db"))
	if err != nil {
		fmt.Println("数据库打开失败！", err.Error())
		return
	}
	defer db.Close()

	conf, err = config.ParseConf()
	if err != nil {
		fmt.Println("配置文件解析失败！", err.Error())
		return
	}

	mh = models.ModelHelper{Db: db, Conf: conf}
	err = mh.InitDB()
	if err != nil {
		fmt.Println("数据库操作失败！", err.Error())
		return
	}

	//
	eventChannels = mh.GetReposChans()
	RunWorkers()

	m := martini.Classic()

	// 静态文件
	m.Use(martini.Static(filepath.Join(runTimeBasePath, "public")))

	m.Get("/", withAuthOrNot(), viewHome)
	m.Group("/new", func(r martini.Router) {
		r.Post("/repos", newRepos)
		r.Post("/hook", newHook)
	}, withAuthOrNot())
	m.Group("/delete", func(r martini.Router) {
		r.Post("/repos", deleteRepos)
		r.Post("/hook", deleteHook)
	}, withAuthOrNot())

	m.Post("/webhook/(?P<adapter_id>[a-zA-Z]+)/(?P<repos_id>[0-9]+)", hookEventHandler)

	m.RunOnAddr(fmt.Sprintf(":%s", conf.Port))
}
