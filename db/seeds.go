package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"sync"
	"time"
)

// We pass error channel through seed functions so as to align sequential nature of
// database seeding.
//
// We choose to explicitly load the channel at each error case instead of
// deferring a panic function, because at current a panic function has a high
// associated time cost (approx +250ns in Go 1.13)
// [https://stackoverflow.com/questions/32541]

func DropGroupUserRolesTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("drop group user roles table")

	queryStr := fmt.Sprintf("DROP TABLE IF EXISTS %s", ConciergeTables.GroupUserRoles)
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func DropGroupUsersTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("drop group users table")

	queryStr := fmt.Sprintf("DROP TABLE IF EXISTS %s", ConciergeTables.GroupUsers)
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func DropUsersTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("drop users table")

	queryStr := fmt.Sprintf("DROP TABLE IF EXISTS %s", ConciergeTables.Users)
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func DropRolesTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("drop roles table")

	queryStr := fmt.Sprintf("DROP TABLE IF EXISTS %s", ConciergeTables.Roles)
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func DropGroupsTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("drop groups table")

	queryStr := fmt.Sprintf("DROP TABLE IF EXISTS %s", ConciergeTables.Groups)
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func DropRegisteredProcessesTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("drop registered processes table")

	queryStr := fmt.Sprintf("DROP TABLE IF EXISTS %s", ConciergeTables.RegisteredProcesses)
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func DropRegisteredProcessPermissionsTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("drop registered process permissions table")

	queryStr := fmt.Sprintf("DROP TABLE IF EXISTS %s", ConciergeTables.RegisteredProcessPermissions)
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func DropRunningProcessesTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("drop running processes table")

	queryStr := fmt.Sprintf("DROP TABLE IF EXISTS %s", ConciergeTables.RunningProcesses)
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func CreateUsersTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("create users table")

	queryStr := `
        CREATE TABLE IF NOT EXISTS ` +
		ConciergeTables.Users +
		` (
          uid SERIAL PRIMARY KEY,
          username VARCHAR(255) UNIQUE,
          email VARCHAR(255) UNIQUE,
          email_verified BOOLEAN,
          date_created TIMESTAMPTZ,
          last_login TIMESTAMPTZ,
          password VARCHAR(255) NOT NULL
        )
        `
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func CreateRolesTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("create roles table")

	queryStr := `
		CREATE TABLE IF NOT EXISTS ` +
		ConciergeTables.Roles +
		` (
		  rid SERIAL PRIMARY KEY,
		  name VARCHAR(255) UNIQUE
		)
		`
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func CreateGroupsTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("create groups table")

	queryStr := `
        CREATE TABLE IF NOT EXISTS ` +
		ConciergeTables.Groups +
		` (
          gid SERIAL PRIMARY KEY,
          name VARCHAR(255) UNIQUE
        )
        `
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func CreateGroupUsersTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("create group users table")

	queryStr := `
        CREATE TABLE IF NOT EXISTS ` +
		ConciergeTables.GroupUsers +
		` (
          uid SERIAL NOT NULL,
          gid SERIAL NOT NULL,
          FOREIGN KEY (uid) REFERENCES ` +
		ConciergeTables.Users +
		` (uid),
          FOREIGN KEY (gid) REFERENCES ` +
		ConciergeTables.Groups +
		` (gid)
        );
        `
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func CreateGroupUserRolesTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("create group user roles table")

	queryStr := `
        CREATE TABLE IF NOT EXISTS ` +
		ConciergeTables.GroupUserRoles +
		` (
        uid SERIAL NOT NULL,
        gid SERIAL NOT NULL,
        rid SERIAL NOT NULL,
        FOREIGN KEY (uid) REFERENCES ` +
		ConciergeTables.Users + ` (uid),
        FOREIGN KEY (gid) REFERENCES ` +
		ConciergeTables.Groups + ` (gid),
        FOREIGN KEY (rid) REFERENCES ` +
		ConciergeTables.Roles + ` (rid)
        );
        `
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func CreateRegisteredProcessesTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("create registered processes table")

	queryStr := `
        CREATE TABLE IF NOT EXISTS ` +
		ConciergeTables.RegisteredProcesses +
		` (
          rpid SERIAL PRIMARY KEY,
          creator_uid SERIAL NOT NULL,
          name VARCHAR(255) UNIQUE,
          run_command VARCHAR(255),
          kill_command VARCHAR(255),
          date_created TIMESTAMPTZ,
          FOREIGN KEY (creator_uid) REFERENCES ` +
		ConciergeTables.Users + ` (uid)
        );
        `
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func CreateRegisteredProcessPermissionsTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("create registered process permissions table")

	queryStr := `
        CREATE TABLE IF NOT EXISTS ` +
		ConciergeTables.RegisteredProcessPermissions +
		` (
        rpid SERIAL NOT NULL,
        gid SERIAL NOT NULL,
        rid SERIAL NOT NULL,
        rwx BIT(3),
        FOREIGN KEY (rpid) REFERENCES ` +
		ConciergeTables.RegisteredProcesses + ` (rpid),
        FOREIGN KEY (gid) REFERENCES ` +
		ConciergeTables.Groups + ` (gid),
        FOREIGN KEY (rid) REFERENCES ` +
		ConciergeTables.Roles + ` (rid)
        );
        `
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func CreateRunningProcessesTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("create running processes table")

	queryStr := `
        CREATE TABLE IF NOT EXISTS ` +
		ConciergeTables.RunningProcesses +
		` (
        name VARCHAR(255) UNIQUE,
        pid BIGINT NOT NULL,
        runner_uid SERIAL NOT NULL,
        gid SERIAL NOT NULL,
        rpid SERIAL NOT NULL,
        FOREIGN KEY (runner_uid) REFERENCES ` +
		ConciergeTables.Users + ` (uid),
        FOREIGN KEY (gid) REFERENCES ` +
		ConciergeTables.Groups + ` (gid),
        FOREIGN KEY (rpid) REFERENCES ` +
		ConciergeTables.RegisteredProcesses + ` (rpid)
        );
        `
	_, err := db.Query(queryStr)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func SeedUsersTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("seed users table")

	emailVerified := false
	dateCreated := pq.FormatTimestamp(time.Now())
	lastLogin := pq.FormatTimestamp(time.Now())
	for i := 0; i < len(InitUsers); i++ {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(InitUsers[i].Password), 8)
		queryStr := `
	         INSERT INTO ` +
			ConciergeTables.Users +
			` (
	           username, email, email_verified, date_created, last_login, password
	         ) VALUES ($1, $2, $3, $4, $5, $6)
			`
		_, err = db.Query(
			queryStr,
			InitUsers[i].Name,
			InitUsers[i].Email,
			emailVerified,
			dateCreated,
			lastLogin,
			passwordHash,
		)
		if err != nil {
			errorChan <- err
			return
		}
	}
	errorChan <- nil
}

func SeedRolesTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("seed roles table")

	queryStr := `
		INSERT INTO ` +
		ConciergeTables.Roles +
		` (name)
		VALUES ($1)
		`
	_, err := db.Query(
		queryStr,
		InitConciergeRoles.User,
	)
	if err != nil {
		errorChan <- err
		return
	}

	queryStr = `
		INSERT INTO ` +
		ConciergeTables.Roles +
		` (name)
		VALUES ($1)
		`
	_, err2 := db.Query(
		queryStr,
		InitConciergeRoles.Admin,
	)
	if err2 != nil {
		errorChan <- err2
		return
	}
	errorChan <- nil
}

func SeedGroupsTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("seed groups table")

	queryStr := `
		INSERT INTO ` +
		ConciergeTables.Groups +
		` (name)
		VALUES ($1)
		`
	_, err := db.Query(
		queryStr,
		InitConciergeGroups.Site,
	)
	if err != nil {
		errorChan <- err
		return
	}
	errorChan <- nil
}

func SeedGroupUsersTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("seed group users table")

	var res1 *sql.Rows
	var res2 *sql.Rows
	var err1, err2, err3 error = nil, nil, nil
	var gid, uid int
	var queryStr string

	for i := 0; i < len(InitUsers); i++ {
		queryStr = `
			SELECT u.uid
			FROM ` +
			ConciergeTables.Users +
			` u
			WHERE u.username = $1
			`
		res1, err1 = db.Query(
			queryStr,
			InitUsers[i].Name,
		)
		if err1 != nil {
			errorChan <- err1
			return
		} else if res1.Next() {
			if err1 = res1.Scan(&uid); err1 != nil {
				errorChan <- err1
				return
			}
		} else {
			errString := fmt.Sprintf("No uid found for user %s", InitUsers[i].Name)
			errorChan <- errors.New(errString)
			return
		}

		for j := 0; j < len(InitUsers[i].Roles); j++ {
			queryStr = `
				SELECT g.gid
				FROM ` +
				ConciergeTables.Groups +
				` g
				WHERE g.name = $1
				`
			res2, err2 = db.Query(
				queryStr,
				InitUsers[i].Roles[j].Group,
			)
			if err2 != nil {
				errorChan <- err2
				return
			} else if res2.Next() {
				if err2 = res2.Scan(&gid); err2 != nil {
					errorChan <- err2
					return
				}
			} else {
				errString := fmt.Sprintf(
					"No gid found for group %s",
					InitUsers[i].Name,
					InitUsers[i].Roles[j].Group,
				)
				errorChan <- errors.New(errString)
				return
			}

			queryStr = `
				INSERT INTO ` +
				ConciergeTables.GroupUsers +
				`	(uid, gid)
				VALUES
					($1, $2)
				`
			_, err3 = db.Query(
				queryStr,
				uid,
				gid,
			)
			if err3 != nil {
				errorChan <- err3
				return
			}
		}
	}
	errorChan <- nil
}

func SeedGroupUserRolesTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("seed group user roles table")

	var res1 *sql.Rows
	var res2 *sql.Rows
	var res3 *sql.Rows
	var err1, err2, err3, err4 error = nil, nil, nil, nil
	var gid, uid, rid int
	var queryStr string

	for i := 0; i < len(InitUsers); i++ {
		queryStr = `
			SELECT u.uid
			FROM ` +
			ConciergeTables.Users +
			` u
			WHERE u.username = $1
			`
		res1, err1 = db.Query(
			queryStr,
			InitUsers[i].Name,
		)
		if err1 != nil {
			errorChan <- err1
			return
		} else if res1.Next() {
			if err1 = res1.Scan(&uid); err1 != nil {
				errorChan <- err1
				return
			}
		} else {
			errString := fmt.Sprintf("No uid found for user %s", InitUsers[i].Name)
			errorChan <- errors.New(errString)
			return
		}

		for j := 0; j < len(InitUsers[i].Roles); j++ {
			role := InitUsers[i].Roles[j]
			queryStr = `
				SELECT g.gid
				FROM ` +
				ConciergeTables.Groups +
				` g
				WHERE g.name = $1
				`
			res2, err2 = db.Query(
				queryStr,
				role.Group,
			)

			if err2 != nil {
				errorChan <- err2
				return
			} else if res2.Next() {
				if err2 = res2.Scan(&gid); err2 != nil {
					errorChan <- err2
					return
				}
			} else {
				errString := fmt.Sprintf(
					"No gid found for group %s",
					InitUsers[i].Name,
					role.Group,
				)
				errorChan <- errors.New(errString)
				return
			}

			queryStr = `
				SELECT r.rid
				FROM ` +
				ConciergeTables.Roles +
				` r
				WHERE r.name = $1
				`
			res3, err3 = db.Query(
				queryStr,
				role.Role,
			)
			if err3 != nil {
				errorChan <- err3
				return
			} else if res3.Next() {
				if err3 = res3.Scan(&rid); err3 != nil {
					errorChan <- err3
					return
				}
			} else {
				errString := fmt.Sprintf(
					"No rid found for role %s",
					InitUsers[i].Name,
					role.Role,
				)
				errorChan <- errors.New(errString)
				return
			}

			queryStr = `
				INSERT INTO ` +
				ConciergeTables.GroupUserRoles +
				`	(uid, gid, rid)
				Values
					($1, $2, $3)
				`
			_, err4 = db.Query(
				queryStr,
				uid,
				gid,
				rid,
			)
			if err4 != nil {
				errorChan <- err4
				return
			}
		}
	}
	errorChan <- nil
}

func SeedRegisteredProcessesTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("seed registered processes table")
	errorChan <- nil
}

func SeedRegisteredProcessPermissionsTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("seed registered process permissions table")
	errorChan <- nil
}

func SeedRunningProcessesTable(db *sql.DB, wg *sync.WaitGroup, errorChan chan error) {
	defer wg.Done()
	fmt.Println("seed running processes table")
	errorChan <- nil
}
