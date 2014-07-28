package main

import (
	"log"
	"net/http"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

var masterAbsPath string

func genResponseStr(status string, message string) []byte {
	resp := Response {
		Status: status,
		Msg: message,
	}
	responseContent, _ := json.MarshalIndent(resp, "", "    ")
	return responseContent
}

func HookHandler(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	// 非POST请求的处理
	if req.Method != "POST" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write(genResponseStr("Failed", "请使用POST请求"));
		return
	}
	// POST请求的处理
	var prb PushRequestBody
	eventDecoder := json.NewDecoder(req.Body)
	err := eventDecoder.Decode(&prb)
	if err != nil {
		log.Fatal("请求内容非JSON格式：", err)
		w.Write(genResponseStr("Failed", "请求内容非JSON格式！"))
		return
	}
	if prb.Ref != "refs/heads/master" {
		log.Println("非master的push")
		w.Write(genResponseStr("success", "非master的push，所以未做什么具体的操作！"))
		return
	}
	// 获取当前目录，用于切换回来
	pwd, _ := os.Getwd()
	// 切换当前目录到master分支的代码目录
	err = os.Chdir(masterAbsPath)
	if err != nil {
		w.Write(genResponseStr("Error", "系统错误！"))
		return
	}
	// 先清除可能存在的本地变更
	checkoutCmd := exec.Command("git", "checkout", "*")
	output, err := checkoutCmd.Output()
	if err != nil {
		log.Println(err)
		w.Write(genResponseStr("Failed", "清除本地变更失败！"))
		return
	}
	log.Println(string(output))
	// 然后从可能的非master分支切换回master
	checkoutMasterCmd := exec.Command("git", "checkout", "master")
	output, err = checkoutMasterCmd.Output()
	if err != nil {
		log.Println(err)
		w.Write(genResponseStr("Failed", "工作目录切换到master失败"))
		return
	}
	log.Println(string(output))
	// 然后pull
	pullCmd := exec.Command("git", "pull", "origin", "master")
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

func main() {
	md := flag.String("master_dir", "", "master分支的源码目录")
	flag.Parse()
	if *md == "" {
		fmt.Println("请提供参数！")
		flag.PrintDefaults()
		return
	}

	// 获取绝对路径
	masterAbsPath, _ = filepath.Abs(*md)
	_, err := os.Stat(masterAbsPath)
	if err != nil && os.IsNotExist(err) {
		fmt.Println("不存在该目录！")
		return
	}
	http.HandleFunc("/webhook", HookHandler)
	err = http.ListenAndServe(":8799", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
