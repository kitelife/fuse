package adapters

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "adapter_manager"
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

func (gls GitlabStruct) Parse(req *http.Request) (filteredEventData adapter_manager.FilteredEventDataStruct) {
    var prbs GitlabPushRequestBodyStruct
    eventDecoder := json.NewDecoder(req.Body)
    err := eventDecoder.Decode(&prbs)
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
        return
    }

    commitCount := len(prbs.Commits)
    if commitCount == 0 {
        fmt.Println("本次push事件中commit数目为0")
        return
    }
    filteredEventData = adapter_manager.FilteredEventDataStruct {
        ReposRemoteURL: prbs.Repository.Url,
        BranchName: branchParts[branchPartsLength-1],
        LatestCommit: prbs.Commits[commitCount-1].Id,
    }
    return filteredEventData
}

func init() {
    gls := GitlabStruct{id: "gitlab"}
    adapter_manager.AdapterRegister("gitlab", gls)
}
