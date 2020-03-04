package db

import (
	"errors"
	"fmt"
	"regexp"
)

type InitDbGroups struct {
	Site string
}

type InitDbRoles struct {
	Admin string
	User  string
	Pm    string
}

type DbUserRole struct {
	Group string
	Role  string
}

type DbUser struct {
	Name     string
	Email    string
	Password string
	Roles    []DbUserRole
}

type DbUsers []DbUser

type Tables struct {
	Users                        string
	Roles                        string
	Groups                       string
	GroupUsers                   string
	GroupUserRoles               string
	RegisteredProcesses          string
	RegisteredProcessPermissions string
	RunningProcesses             string
}

var InitConciergeGroups InitDbGroups
var InitConciergeRoles InitDbRoles
var InitUsers []DbUser
var ConciergePermissions map[string]string
var ConciergeTables Tables

func SetupModels(
	env string,
	siteAdminName string,
	siteAdminEmail string,
	siteAdminPassword string,
) error {
	InitConciergeGroups = InitDbGroups{
		Site: "site",
	}
	InitConciergeRoles = InitDbRoles{
		User:  "user",
		Admin: "admin",
	}
	ConciergePermissions = map[string]string{
		"r":  `1[01]{2}`,
		"w":  `[01]1[01]`,
		"x":  `[01]{2}1`,
		"rw": `11[01]`,
		"rx": `1[01]1`,
		"wx": `[01]11`,
	}
	if env == "test" {
		InitUsers = []DbUser{
			DbUser{
				Name:     siteAdminName,
				Email:    siteAdminEmail,
				Password: siteAdminPassword,
				Roles: []DbUserRole{
					{Group: InitConciergeGroups.Site, Role: InitConciergeRoles.Admin},
				},
			},
			DbUser{
				Name:     "test1",
				Email:    "test1@test.com",
				Password: "test1test1test1",
				Roles: []DbUserRole{
					{Group: InitConciergeGroups.Site, Role: InitConciergeRoles.User},
				},
			},
		}

		ConciergeTables = Tables{
			Users:                        "testusers",
			Roles:                        "testroles",
			Groups:                       "testgroups",
			GroupUsers:                   "test_group_users",
			GroupUserRoles:               "test_group_user_roles",
			RegisteredProcesses:          "test_registered_processes",
			RegisteredProcessPermissions: "test_registered_process_permissions",
			RunningProcesses:             "test_running_processes",
		}

		return nil
	} else if env == "release" {
		InitUsers = []DbUser{
			DbUser{
				Name:     siteAdminName,
				Email:    siteAdminEmail,
				Password: siteAdminPassword,
				Roles: []DbUserRole{
					{Group: InitConciergeGroups.Site, Role: InitConciergeRoles.Admin},
				},
			},
		}

		ConciergeTables = Tables{
			Users:                        "users",
			Roles:                        "roles",
			Groups:                       "groups",
			GroupUsers:                   "group_users",
			GroupUserRoles:               "group_user_roles",
			RegisteredProcesses:          "registered_processes",
			RegisteredProcessPermissions: "registered_process_permissions",
			RunningProcesses:             "running_processes",
		}

		return nil
	} else {
		errString := fmt.Sprintf("Models not setup, Environment type invalid: %s", env)
		return errors.New(errString)
	}
}

func MatchPermission(regexpString string, bitstring string) bool {
	matched, _ := regexp.MatchString(regexpString, bitstring)
	return matched
}
