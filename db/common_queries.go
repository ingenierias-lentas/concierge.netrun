package db

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
)

func GetUid(username string, db *sql.DB, errorChan chan error, uid *int) {
	queryStr := `
			SELECT u.uid
			FROM ` +
		ConciergeTables.Users + ` u
			WHERE u.username = $1
		`
	res, err := db.Query(queryStr, username)
	if err != nil {
		errorChan <- err
		return
	} else if res.Next() {
		if err = res.Scan(uid); err != nil {
			errorChan <- err
			return
		}
	} else {
		errString := fmt.Sprintf("No uid found for user %s", username)
		errorChan <- errors.New(errString)
		return
	}

	errorChan <- nil
}

func GetGid(groupname string, db *sql.DB, errorChan chan error, gid *int) {
	queryStr := `
		SELECT g.gid
		FROM ` +
		ConciergeTables.Groups + ` g
		WHERE g.name = $1
	`
	res, err := db.Query(queryStr, groupname)
	if err != nil {
		errorChan <- err
		return
	} else if res.Next() {
		if err = res.Scan(gid); err != nil {
			errorChan <- err
			return
		}
	} else {
		errString := fmt.Sprintf("No gid found for group %s", groupname)
		errorChan <- errors.New(errString)
		return
	}

	errorChan <- nil
}

func GetRid(rolename string, db *sql.DB, errorChan chan error, rid *int) {
	queryStr := `
		SELECT r.rid
		FROM ` +
		ConciergeTables.Roles + ` r
		WHERE r.name = $1
	`
	res, err := db.Query(queryStr, rolename)
	if err != nil {
		errorChan <- err
		return
	} else if res.Next() {
		if err = res.Scan(rid); err != nil {
			errorChan <- err
			return
		}
	} else {
		errString := fmt.Sprintf("No rid found for role %s", rolename)
		errorChan <- errors.New(errString)
		return
	}

	errorChan <- nil
}

func GetRpid(processname string, db *sql.DB, errorChan chan error, rpid *int) {
	queryStr := `
		SELECT rp.rpid
		FROM ` +
		ConciergeTables.RegisteredProcesses + ` rp
		WHERE rp.name = $1
	`
	res, err := db.Query(queryStr, processname)
	if err != nil {
		errorChan <- err
		return
	} else if res.Next() {
		if err = res.Scan(rpid); err != nil {
			errorChan <- err
			return
		}
	} else {
		errString := fmt.Sprintf("No rpid found for process %s", processname)
		errorChan <- errors.New(errString)
		return
	}

	errorChan <- nil
}

func GetPid(processname string, db *sql.DB, errorChan chan error, pid *int) {
	queryStr := `
		SELECT p.pid
		FROM ` +
		ConciergeTables.RegisteredProcesses + ` p
		WHERE p.name = $1
	`
	res, err := db.Query(queryStr, processname)
	if err != nil {
		errorChan <- err
		return
	} else if res.Next() {
		if err = res.Scan(pid); err != nil {
			errorChan <- err
			return
		}
	} else {
		errString := fmt.Sprintf("No rpid found for running process %s", processname)
		errorChan <- errors.New(errString)
		return
	}

	errorChan <- nil
}

func IsInGroup(
	username string,
	groupname string,
	db *sql.DB,
	errorChan chan error,
	isInGroup *bool,
) {
	var uid, gid int

	uidErrorChan := make(chan error)
	gidErrorChan := make(chan error)

	defer func() {
		close(uidErrorChan)
		close(gidErrorChan)
	}()

	go GetUid(username, db, uidErrorChan, &uid)
	go GetGid(groupname, db, gidErrorChan, &gid)
	uidErr, gidErr := <-uidErrorChan, <-gidErrorChan
	if uidErr != nil {
		errorChan <- uidErr
		return
	}
	if gidErr != nil {
		errorChan <- gidErr
		return
	}

	queryStr := `
		SELECT u.uid
		FROM ` +
		ConciergeTables.GroupUsers + ` gu
		INNER JOIN ` + ConciergeTables.Users + ` u on u.uid = gu.uid
		WHERE gu.uid = $1 AND gu.gid = $2
	`
	res, err := db.Query(queryStr, uid, gid)
	if err != nil {
		*isInGroup = false
		errorChan <- err
		return
	} else if res.Next() {
		*isInGroup = true
	} else {
		*isInGroup = false
		errString := fmt.Sprintf("User %s not in group %s", username, groupname)
		errorChan <- errors.New(errString)
		return
	}

	errorChan <- nil
}

func IsRole(
	username string,
	groupname string,
	rolename string,
	db *sql.DB,
	errorChan chan error,
	isRole *bool,
) {
	var uid, gid, rid int

	uidErrorChan := make(chan error)
	gidErrorChan := make(chan error)
	ridErrorChan := make(chan error)

	defer func() {
		close(uidErrorChan)
		close(gidErrorChan)
		close(ridErrorChan)
	}()

	go GetUid(username, db, uidErrorChan, &uid)
	go GetGid(groupname, db, gidErrorChan, &gid)
	go GetRid(rolename, db, ridErrorChan, &rid)
	uidErr, gidErr, ridErr := <-uidErrorChan, <-gidErrorChan, <-ridErrorChan
	if uidErr != nil {
		errorChan <- uidErr
		return
	}
	if gidErr != nil {
		errorChan <- gidErr
		return
	}
	if ridErr != nil {
		errorChan <- ridErr
		return
	}

	queryStr := `
		SELECT r.rid
		FROM ` +
		ConciergeTables.GroupUserRoles + ` gur
		INNER JOIN ` + ConciergeTables.Roles + ` r ON r.rid = gur.rid
		WHERE gur.uid = $1 AND gur.gid = $2 AND gur.rid = $3
	`
	res, err := db.Query(queryStr, uid, gid, rid)
	if err != nil {
		*isRole = false
		errorChan <- err
		return
	} else if res.Next() {
		*isRole = true
	} else {
		*isRole = false
		errString := fmt.Sprintf("User %s is not role %s in group %s",
			username, rolename, groupname)
		errorChan <- errors.New(errString)
		return
	}

	errorChan <- nil
}

func HasPermission(
	username string,
	groupname string,
	rolename string,
	processname string,
	permissionname string,
	db *sql.DB,
	errorChan chan error,
	hasPermission *bool,
) {
	var gid, rid, rpid int
	var isRole bool
	var rwx string

	roleErrorChan := make(chan error)
	rpidErrorChan := make(chan error)
	gidErrorChan := make(chan error)
	ridErrorChan := make(chan error)
	defer func() {
		close(roleErrorChan)
		close(rpidErrorChan)
		close(gidErrorChan)
		close(ridErrorChan)
	}()

	go IsRole(username, groupname, rolename, db, roleErrorChan, &isRole)
	roleErr := <-roleErrorChan
	if roleErr != nil {
		errorChan <- roleErr
		return
	}

	go GetGid(groupname, db, gidErrorChan, &gid)
	go GetRid(rolename, db, ridErrorChan, &rid)
	go GetRpid(processname, db, rpidErrorChan, &rpid)
	gidErr, ridErr, rpidErr := <-gidErrorChan, <-ridErrorChan, <-rpidErrorChan
	if gidErr != nil {
		errorChan <- gidErr
		return
	}
	if ridErr != nil {
		errorChan <- ridErr
		return
	}
	if rpidErr != nil {
		errorChan <- rpidErr
		return
	}

	queryStr := `
		SELECT rp.rwx
		FROM ` +
		ConciergeTables.RegisteredProcesses + ` rp
		WHERE rp.gid = $1 AND rp.rid = $2 and rp.rpid = $3
	`
	res, err := db.Query(queryStr, gid, rid, rpid)
	if err != nil {
		*hasPermission = false
		errorChan <- err
		return
	} else if res.Next() {
		if err = res.Scan(&rwx); err != nil {
			errorChan <- err
			return
		}
		*hasPermission = MatchPermission(ConciergePermissions[permissionname], rwx)
		if !(*hasPermission) {
			permissionErrString := fmt.Sprintf(
				"User %s with role %s does not have permission %s for process %s in group %s",
				username, rolename, permissionname, processname, groupname)
			errorChan <- errors.New(permissionErrString)
			return
		}
	} else {
		*hasPermission = false
		errString := fmt.Sprintf(
			"User %s with role %s does not have permission %s for process %s in group %s",
			username, rolename, permissionname, processname, groupname)
		errorChan <- errors.New(errString)
		return
	}

	errorChan <- nil
}
