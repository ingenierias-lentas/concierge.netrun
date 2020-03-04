package server

import (
	"github.com/gin-gonic/gin"
	conciergedb "github.com/ingenierias-lentas/netrun/db"
	"net/http"
)

type CommandVerification struct {
	User    string `json:user`
	Group   string `json:group`
	Role    string `json:role`
	Process string `json:process`
}

func evalPermission(c *gin.Context, permissionStr string, errorStr string) error {
	errorChan := make(chan error, 1)
	db = GetDb()
	defer func() {
		close(errorChan)
	}()

	var canDo bool

	var err error = nil
	var cmdver CommandVerification

	if err = c.ShouldBindJSON(&cmdver); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}
	go conciergedb.HasPermission(
		cmdver.User,
		cmdver.Group,
		cmdver.Role,
		cmdver.Process,
		permissionStr,
		db,
		errorChan,
		&canDo,
	)
	err = <-errorChan
	if err != nil || canDo == false {
		c.JSON(http.StatusUnauthorized, gin.H{"status": errorStr})
		return err
	}

	return nil
}

func CanExecute() gin.HandlerFunc {
	return func(c *gin.Context) {
		// create copy of context for usage in goroutines
		cCopy := c.Copy()
		if err := evalPermission(cCopy, "x", "Need execute permission"); err != nil {
			return
		}
		c.Next()
	}
}

func CanWrite() gin.HandlerFunc {
	return func(c *gin.Context) {
		cCopy := c.Copy()
		if err := evalPermission(cCopy, "w", "Need write permission"); err != nil {
			return
		}
		c.Next()
	}
}

func CanRead() gin.HandlerFunc {
	return func(c *gin.Context) {
		cCopy := c.Copy()
		if err := evalPermission(cCopy, "r", "Need read permission"); err != nil {
			return
		}
		c.Next()
	}
}
