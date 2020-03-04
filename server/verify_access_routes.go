package server

import (
	"database/sql"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	conciergedb "github.com/ingenierias-lentas/netrun/db"
	//"github.com/lib/pq"
	"net/http"
	"strings"
	//"time"
)

type ConciergeTokenClaims struct {
	User string `json:"user"`
	jwt.StandardClaims
}

type AuthCheck struct {
	Token string `header:"authorization"`
}

type UserCheck struct {
	User string `json:user`
}

type RolesCheck struct {
	Roles []string `json:roles`
}

type GroupCheck struct {
	User  string `json:user`
	Group string `json:group`
}

type RoleCheck struct {
	User  string `json:user`
	Group string `json:group`
	Role  string `json:group`
}

type AdminCheck struct {
	User  string `json:user`
	Group string `json:group`
}

func VerifyToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var authCheck AuthCheck
		var userCheck UserCheck
		var err error
		var parsedToken *jwt.Token

		if err = c.ShouldBindHeader(&authCheck); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err = c.ShouldBindJSON(&userCheck); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token := authCheck.Token
		bearerPrefix := "Bearer "
		if strings.HasPrefix(token, bearerPrefix) {
			token = strings.TrimPrefix(token, bearerPrefix)
		}
		parsedToken, err = jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			secret := GetJwtSecret()
			return secret, nil
		})

		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				c.JSON(http.StatusBadRequest, gin.H{"status": "Authentication token malformed"})
				return
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				c.JSON(http.StatusUnauthorized, gin.H{"status": "Authentication token invalid"})
				return
			} else {
				c.JSON(
					http.StatusInternalServerError,
					gin.H{"status": "Could not handle authentication token"},
				)
				return
			}
		} else if !parsedToken.Valid {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Could not handle authentication token"})
			return
		}

		parsedTokenClaims, ok := parsedToken.Claims.(*ConciergeTokenClaims)

		if !(ok && parsedToken.Valid) {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "Authentication token invalid"})
			return
		} else if userCheck.User != parsedTokenClaims.User {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "Authentication token invalid for provided user"})
			return
		}

		c.Next()
	}
}

func CheckRolesExist() gin.HandlerFunc {
	return func(c *gin.Context) {
		var res *sql.Rows
		var err error
		var queryStr string
		var rolesList RolesCheck
		var currRole string
		var foundRole bool
		db = GetDb()

		if err = c.ShouldBindJSON(&rolesList); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		queryStr = `SELECT name FROM ` + conciergedb.ConciergeTables.Roles
		res, err = db.Query(queryStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Error finding roles"})
			return
		}
		for res.Next() {
			if err = res.Scan(&currRole); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "Error checking roles"})
				return
			}
			foundRole = false
			for _, n := range rolesList.Roles {
				if currRole == n {
					foundRole = true
				}
			}
			if !foundRole {
				errStr := fmt.Sprintf("Role %s not found", currRole)
				c.JSON(http.StatusBadRequest, gin.H{"status": errStr})
				return
			}
		}

		c.Next()
	}
}

func CheckGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var groupCheck GroupCheck
		var isInGroup bool
		var err error
		db = GetDb()

		if err = c.ShouldBindJSON(&groupCheck); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		errorChan := make(chan error, 1)

		defer func() {
			close(errorChan)
		}()

		go conciergedb.IsInGroup(groupCheck.User, groupCheck.Group, db, errorChan, &isInGroup)
		err = <-errorChan
		if err != nil {
			errStr := fmt.Sprintf("User %s not in group %s", groupCheck.User, groupCheck.Group)
			c.JSON(http.StatusBadRequest, gin.H{"status": errStr})
			return
		}

		c.Next()
	}
}

func CheckRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		var roleCheck RoleCheck
		var validRole bool
		var err error
		db = GetDb()

		errorChan := make(chan error, 1)

		defer func() {
			close(errorChan)
		}()

		if err = c.ShouldBindJSON(&roleCheck); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		go conciergedb.IsRole(
			roleCheck.User,
			roleCheck.Group,
			roleCheck.Role,
			db,
			errorChan,
			&validRole,
		)
		err = <-errorChan
		if err != nil {
			errStr := fmt.Sprintf(
				"User %s not role %s in group %s",
				roleCheck.User, roleCheck.Role, roleCheck.Group,
			)
			c.JSON(http.StatusBadRequest, gin.H{"status": errStr})
			return
		}

		c.Next()
	}
}

func IsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var adminCheck AdminCheck
		var isAdmin bool
		var err error
		db = GetDb()

		errorChan := make(chan error, 1)

		defer func() {
			close(errorChan)
		}()

		if err = c.ShouldBindJSON(&adminCheck); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		go conciergedb.IsRole(
			adminCheck.User,
			adminCheck.Group,
			conciergedb.InitConciergeRoles.Admin,
			db,
			errorChan,
			&isAdmin,
		)
		err = <-errorChan
		if err != nil {
			errStr := fmt.Sprintf(
				"User %s not %s in group %s",
				adminCheck.User, conciergedb.InitConciergeRoles.Admin, adminCheck.Group,
			)
			c.JSON(http.StatusBadRequest, gin.H{"status": errStr})
			return
		}

		c.Next()
	}
}
