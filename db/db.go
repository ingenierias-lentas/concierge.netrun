package db

import (
	"database/sql"
	//"errors"
	"fmt"
	_ "github.com/lib/pq"
	"sync"
)

var ConciergeDb *sql.DB = nil

func SetupDb(
	user string,
	host string,
	name string,
	password string,
	port int,
	reset bool,
) (*sql.DB, error) {
	var err error = nil

	// Make error channel to pass through seed functions
	errorChan := make(chan error, 3)

	defer func() {
		close(errorChan)
		if err != nil && ConciergeDb != nil {
			ConciergeDb.Close()
		}
	}()

	fmt.Println("Setting up concierge database")
	connStr := fmt.Sprintf("user=%s host=%s dbname=%s password=%s port=%d sslmode=%s",
		user,
		host,
		name,
		password,
		port,
		"disable",
	)

	ConciergeDb, err = sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Drop current db tables
	var DbWaitGroup sync.WaitGroup
	DbWaitGroup.Add(1)

	go DropGroupUserRolesTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	err = <-errorChan
	if err != nil {
		return nil, err
	}

	DbWaitGroup.Add(1)
	go DropGroupUsersTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	err = <-errorChan
	if err != nil {
		return nil, err
	}

	DbWaitGroup.Add(1)
	go DropRunningProcessesTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	err = <-errorChan
	if err != nil {
		return nil, err
	}

	DbWaitGroup.Add(1)
	go DropRegisteredProcessPermissionsTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	err = <-errorChan
	if err != nil {
		return nil, err
	}

	DbWaitGroup.Add(1)
	go DropRegisteredProcessesTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	err = <-errorChan
	if err != nil {
		return nil, err
	}

	DbWaitGroup.Add(3)
	go DropUsersTable(ConciergeDb, &DbWaitGroup, errorChan)
	go DropRolesTable(ConciergeDb, &DbWaitGroup, errorChan)
	go DropGroupsTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	for i := 0; i < 3; i++ {
		err = <-errorChan
		if err != nil {
			return nil, err
		}
	}

	// Create db tables
	DbWaitGroup.Add(3)
	go CreateUsersTable(ConciergeDb, &DbWaitGroup, errorChan)
	go CreateRolesTable(ConciergeDb, &DbWaitGroup, errorChan)
	go CreateGroupsTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	for i := 0; i < 3; i++ {
		err = <-errorChan
		if err != nil {
			return nil, err
		}
	}

	DbWaitGroup.Add(2)
	go CreateGroupUsersTable(ConciergeDb, &DbWaitGroup, errorChan)
	go CreateGroupUserRolesTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	for i := 0; i < 2; i++ {
		err = <-errorChan
		if err != nil {
			return nil, err
		}
	}

	DbWaitGroup.Add(1)
	go CreateRegisteredProcessesTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	err = <-errorChan
	if err != nil {
		return nil, err
	}

	DbWaitGroup.Add(1)
	go CreateRegisteredProcessPermissionsTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	err = <-errorChan
	if err != nil {
		return nil, err
	}

	DbWaitGroup.Add(1)
	go CreateRunningProcessesTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	err = <-errorChan
	if err != nil {
		return nil, err
	}

	// Insert initial data
	DbWaitGroup.Add(3)
	go SeedUsersTable(ConciergeDb, &DbWaitGroup, errorChan)
	go SeedRolesTable(ConciergeDb, &DbWaitGroup, errorChan)
	go SeedGroupsTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	for i := 0; i < 3; i++ {
		err = <-errorChan
		if err != nil {
			return nil, err
		}
	}

	DbWaitGroup.Add(1)
	go SeedGroupUsersTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	err = <-errorChan
	if err != nil {
		return nil, err
	}

	DbWaitGroup.Add(1)
	go SeedGroupUserRolesTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	err = <-errorChan
	if err != nil {
		return nil, err
	}

	DbWaitGroup.Add(1)
	go SeedRegisteredProcessesTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	err = <-errorChan
	if err != nil {
		return nil, err
	}

	DbWaitGroup.Add(1)
	go SeedRegisteredProcessPermissionsTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	err = <-errorChan
	if err != nil {
		return nil, err
	}

	DbWaitGroup.Add(1)
	go SeedRunningProcessesTable(ConciergeDb, &DbWaitGroup, errorChan)
	DbWaitGroup.Wait()
	err = <-errorChan
	if err != nil {
		return nil, err
	}

	return ConciergeDb, nil
}
