package netrun

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	conciergedb "github.com/ingenierias-lentas/netrun/db"
	"github.com/ingenierias-lentas/netrun/server"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"
)

var srv *gin.Engine
var db *sql.DB
var config map[interface{}]interface{}
var configDb map[interface{}]interface{}
var configSite map[interface{}]interface{}
var configSiteAdmin map[interface{}]interface{}
var logger *zap.Logger

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		fmt.Errorf("Could not initialize zap logger")
	}

	defer logger.Sync()

	logger.Info("===Performing test setup===")

	// If this doesn't work with executable, try runtime.Caller(1)
	_, currentFile, _, ok := runtime.Caller(0)
	if ok == false {
		logger.Error("Error reading runtime directory path")
	}
	configFile := path.Join(path.Dir(currentFile), "/concierge_config.yaml")
	config = LoadConciergeConfig(configFile)
	configDb = config["Db"].(map[interface{}]interface{})
	configSite = config["Site"].(map[interface{}]interface{})
	configSiteAdmin = configSite["Admin"].(map[interface{}]interface{})

	adminUser := configSiteAdmin["User"].(string)

	err = conciergedb.SetupModels(
		config["Env"].(string),
		adminUser,
		configSiteAdmin["Email"].(string),
		configSiteAdmin["Password"].(string),
	)
	if err != nil {
		logger.Error("Error setting up databse models", zap.String("error", err.Error()))
	}

	db, err = conciergedb.SetupDb(
		configDb["User"].(string),
		configDb["Host"].(string),
		configDb["Name"].(string),
		configDb["Password"].(string),
		configDb["Port"].(int),
		config["Reset"].(bool),
	)
	if err != nil {
		logger.Error("Error setting up concierge database", zap.String("error", err.Error()))
	}

	server.SetLogger(logger)
	server.SetJwtSecret([]byte(config["JwtSecret"].(string)))
	server.SetDb(db)
	srv = server.InitServer()

	os.Exit(m.Run())
}

func TestConfigs(t *testing.T) {
	logger.Info("===Testing configs===")

	// If this doesn't work with executable, try runtime.Caller(1)
	_, currentFile, _, ok := runtime.Caller(0)
	if ok == false {
		t.Errorf("Error reading runtime directory path")
	}
	configFile := path.Join(path.Dir(currentFile), "/concierge_config.yaml")
	config := LoadConciergeConfig(configFile)

	var configEnv string
	configEnv, ok = config["Env"].(string)
	if !ok {
		t.Errorf("Config: Environment must be defined")
	}
	logger.Info("Current configuration ", zap.String("env", configEnv))

	var configDb map[interface{}]interface{}
	configDb, ok = config["Db"].(map[interface{}]interface{})
	if configDb == nil || !ok {
		t.Errorf("Config: Database must be defined")
	}

	var configWhitelistSites []interface{}
	configWhitelistSites, ok = config["WhitelistSites"].([]interface{})
	if configWhitelistSites == nil || !ok {
		t.Errorf("Config: Whitelist sites must be defined")
	}

	var configReset bool
	configReset, ok = config["Reset"].(bool)
	if !ok {
		t.Errorf("Config: Reset must be defined")
	}
	logger.Info("Will reset concierge environment?", zap.Bool("reset", configReset))

	var configSite map[interface{}]interface{}
	configSite, ok = config["Site"].(map[interface{}]interface{})
	if configSite == nil || !ok {
		t.Errorf("Config: Site must be defined")
	}

	var configSiteAdmin map[interface{}]interface{}
	configSiteAdmin, ok = configSite["Admin"].(map[interface{}]interface{})
	if configSiteAdmin == nil || !ok {
		t.Errorf("Config: Site admin must be defined")
	}

}

func TestCommonQueries(t *testing.T) {
	logger.Info("===Testing common database queries===")

	var adminUid int
	uidErrorChan := make(chan error, 1)
	var uidErr, gidErr, ridErr error = nil, nil, nil

	adminUser := configSiteAdmin["User"].(string)

	logger.Info("Testing db.GetUid()")
	go conciergedb.GetUid(
		adminUser,
		db,
		uidErrorChan,
		&adminUid,
	)
	uidErr = <-uidErrorChan
	if uidErr != nil {
		t.Errorf("Error fetching uid for %s", adminUser)
	}

	logger.Info("Testing db.GetGid()")
	var siteGid int
	gidErrorChan := make(chan error, 1)
	go conciergedb.GetGid(
		conciergedb.InitConciergeGroups.Site,
		db,
		gidErrorChan,
		&siteGid,
	)
	gidErr = <-gidErrorChan
	if gidErr != nil {
		t.Errorf("Error fetching gid for %s", conciergedb.InitConciergeGroups.Site)
	}

	logger.Info("Testing db.GetRid()")
	var userRid int
	ridErrorChan := make(chan error, 1)
	go conciergedb.GetRid(
		conciergedb.InitConciergeRoles.User,
		db,
		ridErrorChan,
		&userRid,
	)
	ridErr = <-ridErrorChan
	if ridErr != nil {
		t.Errorf("Error fetching rid for %s", conciergedb.InitConciergeRoles.User)
	}

	logger.Info("Testing db.IsInGroup()")
	var isInGroup bool
	isInGroupErrorChan := make(chan error, 1)
	go conciergedb.IsInGroup(
		adminUser,
		conciergedb.InitConciergeGroups.Site,
		db,
		isInGroupErrorChan,
		&isInGroup,
	)
	isInGroupErr := <-isInGroupErrorChan
	if isInGroupErr != nil {
		t.Errorf(
			"Error finding %s in group %s",
			adminUser,
			conciergedb.InitConciergeGroups.Site,
		)

	}

	logger.Info("Testing db.IsRole()")
	var isRole bool
	isRoleErrorChan := make(chan error, 1)
	go conciergedb.IsRole(
		adminUser,
		conciergedb.InitConciergeGroups.Site,
		conciergedb.InitConciergeRoles.Admin,
		db,
		isRoleErrorChan,
		&isRole,
	)
	isRoleErr := <-isRoleErrorChan
	if isRoleErr != nil {
		t.Errorf(
			"Error finding role %s for user %s in group %s",
			conciergedb.InitConciergeRoles.Admin,
			adminUser,
			conciergedb.InitConciergeGroups.Site,
		)
	}
}

func TestPingRoute(t *testing.T) {
	logger.Info("===Testing ping route===")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	srv.ServeHTTP(w, req)
	if w.Body.String() != "pong" {
		t.Errorf("/ping response = %s ; want pong", w.Body.String())
	}
	if w.Code != 200 {
		t.Errorf("/ping code = %d ; want 200", w.Code)
	}
}

func TestRoutes(t *testing.T) {
	var signinRes server.SigninRes
	user := "bigtest1"
	email := "big@test.com"
	password := "onetwothreefourfive"
	group := ""
	role := ""
	commandname := "testcommand"
	runcommand := "sleep 1000"
	killcommand := ""

	t.Run("Routes=/access", func(t *testing.T) {
		logger.Info("===Testing access routes===")

		w := httptest.NewRecorder()
		reqBody, err := json.Marshal(map[string]string{
			"user":     user,
			"email":    email,
			"password": password,
		})
		if err != nil {
			t.Errorf(err.Error())
		}
		req, _ := http.NewRequest("POST", "/access/signup", bytes.NewBuffer(reqBody))
		srv.ServeHTTP(w, req)
		if w.Code != 200 {
			logger.Error(
				"Incorrect response from /access/signup",
				zap.String("body", w.Body.String()),
			)
			t.Errorf("/access/signup response = %d ; want 200", w.Code)
		}

		w = httptest.NewRecorder()
		reqBody, err = json.Marshal(map[string]string{
			"user":     user,
			"password": password,
		})
		if err != nil {
			t.Errorf(err.Error())
		}
		req, _ = http.NewRequest("POST", "/access/signin", bytes.NewBuffer(reqBody))
		srv.ServeHTTP(w, req)
		fmt.Println(w.Code)
		if w.Code != 200 {
			logger.Error(
				"Incorrect response from /access/signin",
				zap.String("body", w.Body.String()),
			)
			t.Errorf("/access/signin response = %d ; want 200", w.Code)
		}
		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &signinRes)
		if err != nil {
			t.Errorf("Error decoding signin response")
		}

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/access/signout", nil)
		srv.ServeHTTP(w, req)
		if w.Code != 200 {
			logger.Error(
				"Incorrect response from /access/signout",
				zap.String("body", w.Body.String()),
			)
			t.Errorf("/access/signin response = %d ; want 200", w.Code)
		}
	})

	t.Run("Routes=/command", func(t *testing.T) {
		logger.Info("===Testing command routes===")

		reqBody, err := json.Marshal(map[string]string{
			"user":        user,
			"group":       group,
			"role":        role,
			"commandname": commandname,
			"runcommand":  runcommand,
			"killcommand": killcommand,
		})
		if err != nil {
			t.Errorf(err.Error())
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/command/newcommand", bytes.NewBuffer(reqBody))
		req.Header.Set("authorization", "Bearer "+signinRes.AccessToken)
		srv.ServeHTTP(w, req)
		if w.Code != 200 {
			logger.Error(
				"Incorrect response from /command/newcommand",
				zap.String("body", w.Body.String()),
			)
			t.Errorf("/command/newcommand response = %d ; want 200", w.Code)
		}

		/*

			req, _ := http.NewRequest("POST", "/command/deletecommand", nil)
			srv.ServeHTTP(w, req)

			type NewCommandBody struct {
				User        string `json:user`
				Group       string `json:group`
				CommandName string `json:commandname`
				RunCommand  string `json:runcommand`
				KillCommand string `json:killcommand`
			}

			type DeleteCommandBody struct {
				CommandName string `json:commandname`
			}

			type CommandVerification struct {
				User    string `json:user`
				Group   string `json:group`
				Role    string `json:role`
				Process string `json:process`
			}
		*/
	})
}
