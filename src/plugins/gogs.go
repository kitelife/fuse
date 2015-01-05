package plugins

// 待测试

/*
{
    "secret": "",
    "ref": "refs/heads/master",
    "commits": [
        {
            "id": "5f69e7cedd45fcce5ea8f3116e9e20f15e90dafb",
            "message": "hi\n",
            "url": "http://localhost:3000/unknwon/macaron/commit/5f69e7cedd45fcce5ea8f3116e9e20f15e90dafb",
            "author": {
                "name": "Unknwon",
                "email": "joe2010xtmf@163.com",
                "username": "Unknwon"
            }
        }
    ],
    "repository": {
        "id": 1,
        "name": "macaron",
        "url": "http://localhost:3000/unknwon/macaron",
        "description": "",
        "website": "",
        "watchers": 1,
        "owner": {
            "name": "Unknwon",
            "email": "joe2010xtmf@163.com",
            "username": "Unknwon"
        },
        "private": false
    },
    "pusher": {
        "name": "Unknwon",
        "email": "joe2010xtmf@163.com",
        "username": "unknwon"
    },
    "before": "f22f45d79a2ff050f0250a7df41f4944e6591853",
    "after": "5f69e7cedd45fcce5ea8f3116e9e20f15e90dafb",
    "compare_url": "http://localhost:3000/unknwon/macaron/compare/f22f45d79a2ff050f0250a7df41f4944e6591853...5f69e7cedd45fcce5ea8f3116e9e20f15e90dafb"
}
*/

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "plugin_manager"
)

type GogsStruct struct {
    id string
}

type GogsCommitAuthorInfoStruct struct {
    Name  string
    Email string
    Username string
}

type GogsReposOwnerInfoStruct GogsCommitAuthorInfoStruct
type GogsPusherInfoStruct GogsCommitAuthorInfoStruct

type GogsCommitInfoStruct struct {
    Id        string
    Message   string
    Url       string
    Author    GogsCommitAuthorInfoStruct
}

type GogsReposInfoStruct struct {
    Id          int
    Name        string
    Url         string
    Description string
    Website     string
    Watchers    int
    Owner       GogsReposOwnerInfoStruct
    Private     bool
}

type GogsPushRequestBodyStruct struct {
    Secret          string
    Ref             string
    Commits         []GogsCommitInfoStruct
    Repository      GogsReposInfoStruct
    Pusher          GogsPusherInfoStruct
    Before          string
    After           string
    Compare_url     string
}

func (gogs GogsStruct) Parse(req *http.Request) (reposRemoteURL string, branchName string) {
    var prbs GogsPushRequestBodyStruct
    eventDecoder := json.NewDecoder(req.Body)
    err := eventDecoder.Decode(&prbs)
    if err != nil {
        return "", ""
    }
    // 记录日志
    // reqBodyStr, _ := json.MarshalIndent(prb, "", "    ")
    // log.Println(string(reqBodyStr))

    branchParts := strings.Split(prbs.Ref, "/")
    branchPartsLength := len(branchParts)
    if branchPartsLength == 0 {
        fmt.Println("请求内容中分支不正确！", prbs.Ref)
        return "", ""
    }

    branchName = branchParts[branchPartsLength-1]
    reposRemoteURL = prbs.Repository.Url
    // 这里的reposRemoteURL是需要的远程仓库的地址么？
    return reposRemoteURL, branchName
}

func init() {
    gogs := GogsStruct{id: "gogs"}
    plugin_manager.PluginRegister("gogs", gogs)
}
