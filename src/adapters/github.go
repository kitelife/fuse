package adapters

// 待测试

// push事件的JSON请求体结构参见：https://developer.github.com/v3/activity/events/types/#pushevent

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "adapter_manager"
)

type GithubStruct struct {
    id string
}

type GithubCommitAuthorInfoStruct struct {
    Name string
    Email string
    Username string
}

type GithubCommitterInfoStruct GithubCommitAuthorInfoStruct

type GithubCommitInfoStruct struct {
    Id string
    Distinct bool
    Message string
    Timestamp string
    Url string
    Author GithubCommitAuthorInfoStruct
    Committer GithubCommitterInfoStruct
    Added []string
    Removed []string
    Modified []string
}

type GithubHeadCommitInfoStruct GithubCommitInfoStruct

type GithubReposOwnerInfoStruct struct {
    Name string
    Email string
}

type GithubReposInfoStruct struct {
    Id int
    Name string
    Full_name string
    Owner GithubReposOwnerInfoStruct
    Private bool
    Html_url string
    Description string
    Fork bool
    Url string
    Forks_url string
    Keys_url string
    Collaborators_url string
    Teams_url string
    Hooks_url string
    Issue_events_url string
    Events_url string
    Assignees_url string
    Branches_url string
    Tags_url string
    Blobs_url string
    Git_tags_url string
    Git_refs_url string
    Trees_url string
    Statuses_url string
    Languages_url string
    Stargazers_url string
    Contributors_url string
    Subscribers_url string
    Subscription_url string
    Commits_url string
    Git_commits_url string
    Comments_url string
    Issue_comment_url string
    Contents_url string
    Compare_url string
    Merges_url string
    Archive_url string
    Downloads_url string
    Issues_url string
    Pulls_url string
    Milestones_url string
    Notifications_url string
    Labels_url string
    Releases_url string
    Created_at int
    Updated_at string
    Pushed_at int
    Git_url string
    Ssh_url string
    Clone_url string
    Svn_url string
    Homepage string
    Size int
    Stargazers_count int
    Watchers_count int
    Language string
    Has_issues bool
    Has_downloads bool
    Has_wiki bool
    Has_pages bool
    Forks_count int
    Mirror_url string
    Open_issues_count int
    Forks int
    Open_issues int
    Watchers int
    Default_branch string
    Stargazers int
    Master_branch string
}

type GithubPusherInfoStruct GithubReposOwnerInfoStruct

type GithubSenderInfoStruct struct {
    Login string
    Id int
    Avatar_url string
    Gravatar_id string
    Url string
    Html_url string
    Followers_url string
    Following_url string
    Gists_url string
    Starred_url string
    Subscriptions_url string
    Organizations_url string
    Repos_url string
    Events_url string
    Received_events_url string
    Type string
    Site_admin bool
}

type GithubPushRequestBodyStruct struct {
    Ref string
    Before string
    After string
    Created bool
    Deleted bool
    Forced bool
    Base_ref string
    Compare string
    Commits []GithubCommitInfoStruct
    Head_commit GithubHeadCommitInfoStruct
    Repository GithubReposInfoStruct
    Pusher GithubPusherInfoStruct
    Sender GithubSenderInfoStruct
}

func (github GithubStruct) Parse(req *http.Request) (reposRemoteURL string, branchName string) {
    var prbs GithubPushRequestBodyStruct
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
    reposRemoteURL = prbs.Repository.Git_url
    // 这里的reposRemoteURL是需要的远程仓库的地址么？
    return reposRemoteURL, branchName
}

func init() {
    github := GithubStruct{id: "github"}
    adapter_manager.AdapterRegister("github", github)
}
