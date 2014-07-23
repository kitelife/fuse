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

func HookHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World!\n"));
	} else {
		contentContainer := make([]byte, req.ContentLength)
		_, err := req.Body.Read(contentContainer)
		if err != nil {
			log.Fatal("Read request body: ", err)
			w.Write([]byte("系统错误！"))
		}
		var prb PushRequestBody
		err = json.Unmarshal(contentContainer, &prb)
		if err != nil {
			log.Fatal("请求内容非JSON格式：", err)
			w.Write([]byte("请求内容非JSON格式！"))
			return
		}
		if prb.Ref == "refs/heads/master" {
			// 获取当前目录，用于切换回来
			pwd, _ := os.Getwd()
			// 切换当前目录到master分支的代码目录
			err := os.Chdir(masterAbsPath)
			if err != nil {
				w.Write([]byte("系统错误！"))
				return
			}
			// 先清除可能存在的本地变更
			checkoutCmd := exec.Command("git", "checkout", "*")
			output, err := checkoutCmd.Output()
			if err != nil {
				log.Println(err)
			}
			log.Println(string(output))
			// 然后从可能的非master分支切换回master
			checkoutMasterCmd := exec.Command("git", "checkout", "master")
			output, err = checkoutMasterCmd.Output()
			if err != nil {
				log.Println(err)
			}
			log.Println(string(output))
			// 然后pull
			pullCmd := exec.Command("git", "pull", "origin", "master")
			output, err = pullCmd.Output()
			if err != nil {
				log.Println(err)
			}
			log.Println(string(output))
			// 切换回原工作目录
			os.Chdir(pwd)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := Response {
			Status: "success",
			Msg: "成功！",
		}
		responseContent, _ := json.MarshalIndent(resp, "", "    ")
		w.Write([]byte(responseContent))
	}
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
