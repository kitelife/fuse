package main 

import (
    "database/sql"
)

func CheckReposNameExists(db *sql.DB, reposName string) (bool, error) {
    targetSQL := `SELECT COUNT(*) FROM repos WHERE repos_name="?"`
    rows, err := db.Query(targetSQL, reposName)
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
}

func SaveNewRepos(reposType string, reposName string, reposRemote string) error {
    targetSQL = `INSERT INTO repos (repos_name, repos_remote, repos_type) VALUES ("?", "?", "?")`
    if _, err := db.Exec(targetSQL, reposName, reposRemote, reposType); err != nil {
        return err
    }
    return nil
}