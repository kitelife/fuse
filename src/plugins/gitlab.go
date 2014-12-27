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
	id   string
	Host string
}

func (*Gitlab) Parse(req *http.Request) (reposRemoteURL string, branchName string) {
	var prb GitlabPushRequestBody
	eventDecoder := json.NewDecoder(req.Body)
	err := eventDecoder.Decode(&prb)
	if err != nil {
		return "", ""
	}
	// 记录日志
	// reqBodyStr, _ := json.MarshalIndent(prb, "", "    ")
	// log.Println(string(reqBodyStr))

	branchParts := strings.Split(prb.Ref, "/")
	branchPartsLength = len(branchParts)
	if branchPartsLength == 0 {
		log.Fatalln("请求内容中分支不正确！", prb.Ref)
		w.Write(genResponseStr("Failed", "请求内容不合法！"))
		return
	}

	branchName := branchParts[branchPartsLength-1]
	reposRemoteURL := prb.Repository.Url
	return reposRemoteURL, branchName
}

func init() {
	gl := Gitlab{id: "gitlab"}
	PluginRegister(gl)
}
