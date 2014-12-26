package plugins

import (
    "net/http"
)

type GitlabReposInfo struct {
    Name        string
    Url         string
    Description string
    Homepage    string
}

type GitlabCommitAuthorInfo struct {
    Name  string
    Email string
}

type GitlabCommitInfo struct {
    Id        string
    Message   string
    Timestamp string
    Url       string
    Author    CommitAuthorInfo
}

type GitlabPushRequestBody struct {
    Before     string
    After      string
    Ref        string
    User_id    int
    User_name  string
    Project_id int
    Repository ReposInfo
    Commits    []CommitInfo
}

type Gitlab struct {
    id string
    Host string
    EventData GitlabPushRequestBody
}

func (*Gitlab)IsMe(req *http.Request) bool {
    
}

func (*Gitlab)Parse() (reposRemoteURL string, branchName string) {
    
}

func init() {
    gl := Gitlab{id: "gitlab"}
    PluginRegister(gl)
}