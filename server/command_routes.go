package server

import (
	"github.com/gin-gonic/gin"
	conciergedb "github.com/ingenierias-lentas/netrun/db"
	"github.com/lib/pq"
	"net/http"
	"time"
)

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

func NewCommand(c *gin.Context) {
	var err error = nil
	var cmd NewCommandBody
	var uid, gid, rid, rpid int
	var queryStr string
	uidErrorChan := make(chan error)
	gidErrorChan := make(chan error)
	ridErrorChan := make(chan error)
	rpidErrorChan := make(chan error)
	db = GetDb()

	defer func() {
		close(uidErrorChan)
		close(gidErrorChan)
		close(ridErrorChan)
		close(rpidErrorChan)
	}()

	if err = c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	go conciergedb.GetUid(cmd.User, db, uidErrorChan, &uid)
	go conciergedb.GetGid(cmd.Group, db, gidErrorChan, &gid)
	go conciergedb.GetRid(conciergedb.InitConciergeRoles.Admin, db, ridErrorChan, &rid)
	uidErr, gidErr, ridErr := <-uidErrorChan, <-gidErrorChan, <-ridErrorChan
	if uidErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Cannot find user"})
		return
	}
	if gidErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Cannot find group"})
		return
	}
	if ridErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Cannot find role"})
		return
	}

	dateCreated := pq.FormatTimestamp(time.Now())

	queryStr = `
		INSERT INTO ` +
		conciergedb.ConciergeTables.RegisteredProcesses + `
		  (creator_uid, name, run_command, kill_command, date_created)
		VALUES ($1, $2, $3, $4, $5)
		`

	_, err = db.Query(
		queryStr,
		uid,
		cmd.CommandName,
		cmd.RunCommand,
		cmd.KillCommand,
		dateCreated,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Error registering command"})
		return
	}

	go conciergedb.GetRpid(cmd.CommandName, db, rpidErrorChan, &rpid)
	rpidErr := <-rpidErrorChan
	if rpidErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Error checking registered command"})
	}
	permissionLevel := "B111"

	queryStr = `
        INSERT INTO ` +
		conciergedb.ConciergeTables.RegisteredProcessPermissions + `
          (rpid, gid, rid, rwx)
        VALUES ($1, $2, $3, $4)
        `
	_, err = db.Query(
		queryStr,
		rpid,
		gid,
		rid,
		permissionLevel,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Error registering command permissions"})
		return
	}

	c.String(http.StatusOK, "Command created successfully")
}

func DeleteCommand(c *gin.Context) {
	var err error = nil
	var cmd DeleteCommandBody
	var rpid int
	var queryStr string
	rpidErrorChan := make(chan error)
	db = GetDb()

	defer func() {
		close(rpidErrorChan)
	}()

	if err = c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	go conciergedb.GetRpid(cmd.CommandName, db, rpidErrorChan, &rpid)
	rpidErr := <-rpidErrorChan
	if rpidErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Cannot find command"})
		return
	}

	queryStr = `
        DELETE FROM	` +
		conciergedb.ConciergeTables.RegisteredProcessPermissions + `
        WHERE ` + conciergedb.ConciergeTables.RegisteredProcessPermissions + `.rpid = $1
        `
	_, err = db.Query(queryStr, rpid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Error deleting command permissions"})
		return
	}

	queryStr = `
        DELETE FROM ` +
		conciergedb.ConciergeTables.RegisteredProcesses + `
        WHERE ` + conciergedb.ConciergeTables.RegisteredProcesses + `.rpid = $1
        `
	_, err = db.Query(queryStr, rpid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Error deleting command"})
		return
	}

	c.String(http.StatusOK, "Command deleted successfully")
}

func RunCommand(c *gin.Context) {
	//TODO container stuff
	c.String(http.StatusOK, "Command run successfully")
}

func KillCommand(c *gin.Context) {
	//TODO container stuff
	c.String(http.StatusOK, "Command killed successfully")
}

/*
const exec = require('child_process').exec;

const roles = require('../models/roles');
const {db, tables} = require('../db');
const commonQueries = require('../db/commonQueries');

exports.runCommand = async (req, res) => {
  console.log('Processing func -> Run command');

  try {
    let query1 = {
      text:
        `
        SELECT rp.run_command rp.rpid
        FROM ${tables.registered_processes} rp
        WHERE r.name = $1
        `,
      values:
        [
          req.body.processname
        ]
    };
    let res1 = await db.query(query1);
    const child = exec(res1.rows[0].runCommand,
      (err, stdout, stderr) => {
        console.log(`stdout: ${stdout}`);
        console.log(`stderr: ${stderr}`);
        if (err != null) {
          console.log(`exec error: ${err}`);
          res.status(500).send("Error executing run command -> " + err);
        }
    })
    console.log(child);

    let runnerUid = commonQueries.getUid(req.body.username);
    let gid = commonQueries.getGid(req.body.groupname);
    let query2 = {
      text:
        `
        INSERT INTO ${tables.running_processes} p
          (name, pid, runner_uid, gid, rpid)
        VALUES ($1, $2, $3, $4, $5)
        `,
      values:
        [
          req.body.runningProcessname
          child.pid,
          runnerUid,
          gid,
          res1.rows[0].rpid,
        ]
    }
    let res2 = await db.query(query2);

    res.status(200).json({
      "description": "Command run successfully"
    });
  } catch(err) {
    res.status(500).send("Error running command -> " + err);
  }
}

//TODO remove child id from running_processes database
exports.killCommand = async (req, res) => {
  console.log('Processing func -> Kill command');

  try {
    let query1 = {
      text:
        `
        SELECT rp.kill_command
        FROM ${tables.registered_processes} rp
        WHERE r.name = $1
        `,
      values:
        [
          req.body.processname
        ]
    };
    let res1 = await db.query(query1);
    const child = exec(res1.rows[0].killCommand,
      (err, stdout, stderr) => {
        console.log(`stdout: ${stdout}`);
        console.log(`stderr: ${stderr}`);
        if (error != null) {
          console.log(`exec error: ${err}`);
          res.status(500).send("Error executing kill command -> " + err);
        }
    })

    let pid = commonQueries.getPid(req.body.runningProcessname);
    let query2 = {
      text:
        `
        DELETE FROM ${tables.running_processes}
        WHERE ${tables.running_processes}.name = $1
        `,
      values:
        [
          pid
        ]
    }
    let res2 = await db.query(query2);

    res.status(200).json({
      "description": "Command killed successfully"
    });
  } catch(err) {
    res.status(500).send("Error killing command -> " + err);
  }
}
*/
