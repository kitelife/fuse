package adapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/youngsterxyf/fuse/adapter_manager"
	"net/http"
	"strings"
)

type GitlabReposInfoStruct struct {
	Name        string
	Url         string
	Description string
	Homepage    string
}

type GitlabCommitAuthorInfoStruct struct {
	Name  string
	Email string
}

type GitlabCommitInfoStruct struct {
	Id        string
	Message   string
	Timestamp string
	Url       string
	Author    GitlabCommitAuthorInfoStruct
}

type GitlabPushRequestBodyStruct struct {
	Before     string
	After      string
	Ref        string
	User_id    int
	User_name  string
	Project_id int
	Repository GitlabReposInfoStruct
	Commits    []GitlabCommitInfoStruct
}

type GitlabStruct struct {
	id string
}

func (gls GitlabStruct) Parse(req *http.Request) (filteredEventData adapter_manager.FilteredEventDataStruct, err error) {
	var prbs GitlabPushRequestBodyStruct
	eventDecoder := json.NewDecoder(req.Body)
	err = eventDecoder.Decode(&prbs)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// 记录日志
	// reqBodyStr, _ := json.MarshalIndent(prb, "", "    ")
	// log.Println(string(reqBodyStr))

	branchParts := strings.Split(prbs.Ref, "/")
	branchPartsLength := len(branchParts)
	if branchPartsLength == 0 {
		fmt.Println("请求内容中分支不正确！", prbs.Ref)
		return filteredEventData, errors.New("请求内容中分支不正确！")
	}

	commitCount := len(prbs.Commits)
	if commitCount == 0 {
		return filteredEventData, errors.New("本次push事件中commit数目为0")
	}
	filteredEventData = adapter_manager.FilteredEventDataStruct{
		ReposRemoteURL: prbs.Repository.Url,
		BranchName:     branchParts[branchPartsLength-1],
		LatestCommit:   prbs.Commits[commitCount-1].Id,
	}
	return filteredEventData, nil
}

func init() {
	gls := GitlabStruct{id: "gitlab"}
	adapter_manager.AdapterRegister("gitlab", gls)
}
