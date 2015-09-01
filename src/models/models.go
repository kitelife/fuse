package models

import (
	"database/sql"
	"fmt"
	"time"

	"config"
)

type ModelHelper struct {
	Db   *sql.DB
	Conf config.ConfStruct
}

type ResponseStruct struct {
	Status string
	Msg    string
}

type ReposStruct struct {
	ReposID     int
	ReposName   string
	ReposRemote string
	ReposType   string
	WebHookURL  string
}

type HookStruct struct {
	HookID      int
	ReposID     int
	WhichBranch string
	TargetDir   string
	HookStatus  string
	LogContent  string
	UpdatedTime string
}

type Branch2HookMap map[string][]HookStruct

type DBRelatedDataStruct struct {
	ReposStruct
	Hooks []HookStruct
}

type HomePageDataStruct struct {
	AdapterIDList []string
	ReposList     map[int]string
	DBRelatedData []DBRelatedDataStruct
}

type ChanElementStruct struct {
	ReposID      int
	HookID       int
	RemoteURL    string
	BranchName   string
	TargetDir    string
	LatestCommit string
	Mh           ModelHelper
}

type ReposChanMap map[int]chan ChanElementStruct

func (mh ModelHelper) InitDB() (err error) {
	// 如果目标数据表还不存在则创建
	tableRepos := `CREATE TABLE IF NOT EXISTS repos (
        repos_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        repos_name TEXT NOT NULL,
        repos_remote TEXT NOT NULL DEFAULT "",
        repos_type TEXT NOT NULL
    )`
	_, err = mh.Db.Exec(tableRepos)

	if err != nil {
		return err
	}

	tableHooks := `CREATE TABLE IF NOT EXISTS hooks (
        hook_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        repos_id INTEGER NOT NULL,
        which_branch TEXT NOT NULL DEFAULT "master",
        target_dir TEXT NOT NULL,
        hook_status TEXT NOT NULL,
        log_content TEXT NOT NULL,
        updated_time TIMESTAMP,
        FOREIGN KEY(repos_id) REFERENCES repos(repos_id)
    );`
	_, err = mh.Db.Exec(tableHooks)
	if err != nil {
		return err
	}
	return nil
}

func (mh ModelHelper) CheckReposNameExists(db *sql.DB, reposName string) (bool, error) {
	targetSQL := "SELECT COUNT(*) FROM repos WHERE repos_name=?"
	rows, err := mh.Db.Query(targetSQL, reposName)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var count int
		rows.Scan(&count)
		if count == 0 {
			return false, nil
		}
		return true, nil
	}
	return false, nil
}

func (mh ModelHelper) StoreNewRepos(reposType string, reposName string, reposRemote string) (int, error) {
	targetSQL := "INSERT INTO repos (repos_name, repos_remote, repos_type) VALUES (?, ?, ?)"
	oneTrans, err := mh.Db.Begin()
	if err != nil {
		fmt.Println(err.Error())
		return -1, err
	}
	insertResult, err := mh.Db.Exec(targetSQL, reposName, reposRemote, reposType)
	if err != nil {
		fmt.Println(err.Error())
		return -1, err
	}
	newReposID, err := insertResult.LastInsertId()
	if err != nil {
		fmt.Println(err.Error())
		if rollbackErr := oneTrans.Rollback(); rollbackErr != nil {
			fmt.Println(rollbackErr.Error())
		}
		return -1, nil
	}
	if err := oneTrans.Commit(); err != nil {
		fmt.Println(err.Error())
		return -1, err
	}
	return int(newReposID), nil
}

// mh.StoreNewHook(reposID, whichBranch, targetDir, updatedTime)
func (mh ModelHelper) StoreNewHook(reposID int, whichBranch string, targetDir string, updatedTime string) (int, error) {
	hookStatus := "ready"
	logContent := ""
	targetSQL := "INSERT INTO hooks (repos_id, which_branch, target_dir, hook_status, log_content, updated_time) VALUES (?, ?, ?, ?, ?, ?)"
	oneTrans, err := mh.Db.Begin()
	if err != nil {
		fmt.Println(err.Error())
		return -1, err
	}
	insertResult, err := mh.Db.Exec(targetSQL, reposID, whichBranch, targetDir, hookStatus, logContent, updatedTime)
	if err != nil {
		fmt.Println(err.Error())
		if rollbackErr := oneTrans.Rollback(); rollbackErr != nil {
			fmt.Println(rollbackErr.Error())
		}
		return -1, err
	}
	newHookID, err := insertResult.LastInsertId()
	if err != nil {
		fmt.Println(err.Error())
		if rollbackErr := oneTrans.Rollback(); rollbackErr != nil {
			fmt.Println(rollbackErr.Error())
		}
		return -1, nil
	}
	if err := oneTrans.Commit(); err != nil {
		fmt.Println(err.Error())
		return -1, err
	}
	return int(newHookID), nil
}

func (mh ModelHelper) QueryDBForHookHandler() (map[int]ReposStruct, map[int]Branch2HookMap) {
	// 尝试读取数据
	reposDataSQL := "SELECT repos_id, repos_name, repos_remote, repos_type FROM repos"
	reposRows, err := mh.Db.Query(reposDataSQL)
	if err != nil {
		return nil, nil
	}
	defer reposRows.Close()

	hooksDataSQL := "SELECT hook_id, repos_id, which_branch, target_dir FROM hooks"
	hooksRows, err := mh.Db.Query(hooksDataSQL)
	if err != nil {
		return nil, nil
	}
	defer hooksRows.Close()

	reposAll := make(map[int]ReposStruct)
	reposBranch2Hook := make(map[int]Branch2HookMap)

	var reposID int

	var reposName string
	var reposRemote string
	var reposType string
	var targetWebHookURL string
	for reposRows.Next() {
		reposRows.Scan(&reposID, &reposName, &reposRemote, &reposType)

		targetWebHookURL = fmt.Sprintf("%s:%s/webhook/%s/%d", mh.Conf.Host, mh.Conf.Port, reposType, reposID)

		reposAll[reposID] = ReposStruct{ReposID: reposID, ReposName: reposName,
			ReposRemote: reposRemote, ReposType: reposType,
			WebHookURL: targetWebHookURL,
		}
	}

	var hookID int
	var whichBranch string
	var targetDir string
	for hooksRows.Next() {
		hooksRows.Scan(&hookID, &reposID, &whichBranch, &targetDir)
		_, ok := reposBranch2Hook[reposID]
		if ok == false {
			reposBranch2Hook[reposID] = make(Branch2HookMap)
		}
		_, ok = reposBranch2Hook[reposID][whichBranch]
		if ok == false {
			reposBranch2Hook[reposID][whichBranch] = make([]HookStruct, 0, 5)
		}
		reposBranch2Hook[reposID][whichBranch] = append(reposBranch2Hook[reposID][whichBranch], HookStruct{hookID, reposID, whichBranch, targetDir, "", "", ""})
	}
	return reposAll, reposBranch2Hook
}

func (mh ModelHelper) QueryDBForViewHome() (reposList map[int]string, dbRelatedData []DBRelatedDataStruct) {
	reposList = make(map[int]string)

	reposDataSQL := "SELECT repos_id, repos_name, repos_remote, repos_type FROM repos"
	reposRows, err := mh.Db.Query(reposDataSQL)
	if err != nil {
		fmt.Println("数据库查询出错！", err.Error())
		return
	}
	hooksDataSQL := "SELECT hook_id, repos_id, which_branch, target_dir, hook_status, log_content, updated_time FROM hooks"
	hooksRows, err := mh.Db.Query(hooksDataSQL)
	if err != nil {
		fmt.Println("数据库查询出错！", err.Error())
		return
	}

	var reposID int

	var hookID int
	var whichBranch string
	var targetDir string
	var hookStatus string
	var logContent string
	var updatedTime time.Time

	var hooks map[int][]HookStruct = make(map[int][]HookStruct)
	for hooksRows.Next() {
		hooksRows.Scan(&hookID, &reposID, &whichBranch, &targetDir, &hookStatus, &logContent, &updatedTime)
		if _, ok := hooks[reposID]; ok == false {
			// 1024: 每个代码库最多能够1024个分支，也即1024个hook
			hooks[reposID] = make([]HookStruct, 0, 1024)
		}
		hooks[reposID] = append(hooks[reposID], HookStruct{hookID, reposID, whichBranch, targetDir, hookStatus, logContent, updatedTime.Format("2006-01-02 15:04:05")})
	}

	var reposName string
	var reposRemote string
	var reposType string
	var targetWebHookURL string
	for reposRows.Next() {
		reposRows.Scan(&reposID, &reposName, &reposRemote, &reposType)

		targetWebHookURL = fmt.Sprintf("%s:%s/webhook/%s/%d", mh.Conf.Host, mh.Conf.Port, reposType, reposID)

		dbRelatedData = append(dbRelatedData, DBRelatedDataStruct{ReposStruct{reposID, reposName, reposRemote, reposType, targetWebHookURL}, hooks[reposID]})
		reposList[reposID] = reposName
	}
	return reposList, dbRelatedData
}

func (mh ModelHelper) CheckReposIDExist(reposID int) (bool, error) {
	targetSQL := "SELECT COUNT(*) FROM repos WHERE repos_id=?"
	rows, err := mh.Db.Query(targetSQL, reposID)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var count int
		rows.Scan(&count)
		if count == 0 {
			return false, nil
		}
		return true, nil
	}
	return false, nil
}

func (mh ModelHelper) DeleteRepos(reposID int) error {
	targetSQL := "DELETE FROM repos WHERE repos_id=?"
	if _, err := mh.Db.Exec(targetSQL, reposID); err != nil {
		return err
	}
	return nil
}

func (mh ModelHelper) DeleteHook(hookID int) error {
	targetSQL := "DELETE FROM hooks WHERE hook_id=?"
	if _, err := mh.Db.Exec(targetSQL, hookID); err != nil {
		return err
	}
	return nil
}

func (mh ModelHelper) CheckReposHasHook(reposID int) (bool, error) {
	targetSQL := "SELECT COUNT(*) FROM hooks WHERE repos_id=?"
	rows, err := mh.Db.Query(targetSQL, reposID)
	if err != nil {
		return true, err
	}
	defer rows.Close()

	for rows.Next() {
		var count int
		rows.Scan(&count)
		if count == 0 {
			return false, nil
		}
		return true, nil
	}
	return false, nil
}

func (mh ModelHelper) GetHookTargetDir(hookID int) (string, error) {
	targetSQL := "SELECT target_dir FROM hooks WHERE hook_id=?"
	rows, err := mh.Db.Query(targetSQL, hookID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var targetDir string
	for rows.Next() {
		rows.Scan(&targetDir)
		break
	}
	return targetDir, nil
}

func (mh ModelHelper) UpdateLogStatus(hookID int, hookStatus string, logContent string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	targetSQL := "UPDATE hooks SET hook_status=?, log_content=?, updated_time=? WHERE hook_id=?"
	if _, err := mh.Db.Exec(targetSQL, hookStatus, logContent, now, hookID); err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func (mh ModelHelper) GetReposChans() (rbcs ReposChanMap) {
	rbcs = make(ReposChanMap)
	reposSQL := "SELECT repos_id FROM repos"
	rows, err := mh.Db.Query(reposSQL)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var reposID int
	for rows.Next() {
		rows.Scan(&reposID)
		rbcs[reposID] = make(chan ChanElementStruct, mh.Conf.Queue_length)
	}
	return rbcs
}

func (mh ModelHelper) CheckHookIDExist(hookID int) (bool, error) {
	targetSQL := "SELECT COUNT(*) FROM hooks WHERE hook_id=?"
	rows, err := mh.Db.Query(targetSQL, hookID)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var count int
		rows.Scan(&count)
		if count == 0 {
			return false, nil
		}
		return true, nil
	}
	return false, nil
}

func (mh ModelHelper) UpdateHook(hookID int, branchName string, targetDir string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	hookStatus := "ready"
	logContent := ""
	targetSQL := "UPDATE hooks SET which_branch=?, target_dir=?, hook_status=?, log_content=?, updated_time=? WHERE hook_id=?"
	if _, err := mh.Db.Exec(targetSQL, branchName, targetDir, hookStatus, logContent, now, hookID); err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
