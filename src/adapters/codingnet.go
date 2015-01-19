package adapters

/*
{
    "after": "74e9f243b94dc4c9ab4db3d58d0994db0d62da51",  //推送之后的版本号
    "before": "36aa3b2a1aeddb8c05b3eabf606f9e083cfb0cac",  //推送之前的版本号
    "commits": [
        {
            "committer": {  //提交者
                "email": "xxxxx@gmail.com", 
                "name": "超级小胖"
            }, 
            "sha": "74e9f243b94dc4c9ab4db3d58d0994db0d62da51", 
            "short_message": "update hello.md"
        }
    ], 
    "ref": "master",  //当前提交的分支或标签名
    "repository": {
        "description": "", 
        "name": "hello-coding", 
        "url": "https://coding.net/u/jiong/p/hello-coding/git"
    }, 
    "token": "123"
}
*/

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    
    "adapter_manager"
)

type CodingNetCommitterStruct struct {
    Email string
    Name string
}

type CodingNetCommitInfoStruct struct {
    Committer CodingNetCommitterStruct
    Sha string
    Short_message string
}

type CodingNetRepositoryInfoStruct struct {
    Description string
    Name string
    Url string
}

type CodingNetPushRequestBodyStruct struct {
    After string
    Before string
    Commits []CodingNetCommitInfoStruct
    Ref string
    Repository CodingNetRepositoryInfoStruct
    Token string
}

type CodingNetStruct struct {
    id string
}

func (cn CodingNetStruct) Parse(req *http.Request) (filteredEventData adapter_manager.FilteredEventDataStruct, err error){
    var prbs CodingNetPushRequestBodyStruct
    eventDecoder := json.NewDecoder(req.Body)
    err = eventDecoder.Decode(&prbs)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    // 记录日志
    // reqBodyStr, _ := json.MarshalIndent(prb, "", "    ")
    // log.Println(string(reqBodyStr))

    if len(prbs.Ref) == 0 {
        fmt.Println("请求内容中分支不正确！", prbs.Ref)
        return filteredEventData, errors.New("请求内容中分支不正确！")
    }

    commitCount := len(prbs.Commits)
    if commitCount == 0 {
        return filteredEventData, errors.New("本次push事件中commit数目为0")
    }
    filteredEventData = adapter_manager.FilteredEventDataStruct {
        ReposRemoteURL: prbs.Repository.Url,
        BranchName: prbs.Ref,
        LatestCommit: prbs.Commits[commitCount-1].Sha,
    }
    return filteredEventData, nil
}

func init() {
    cn := CodingNetStruct{id: "codingnet"}
    adapter_manager.AdapterRegister("codingnet", cn)
}