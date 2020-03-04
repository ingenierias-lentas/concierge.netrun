package server

import (
	"database/sql"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	conciergedb "github.com/ingenierias-lentas/netrun/db"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type SignupBody struct {
	User     string `json:user`
	Email    string `json:group`
	Password string `json:group`
}

type SigninBody struct {
	User     string `json:user`
	Password string `json:group`
}

type SigninRes struct {
	User        string
	Auth        bool
	AccessToken string
}

func Signup(c *gin.Context) {
	var signupBody SignupBody
	var err error
	var queryStr string
	db = GetDb()

	if err = c.ShouldBindJSON(&signupBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(signupBody.Password), 8)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Could not generate password hash"})
		return
	}
	emailVerified := false
	nowTime := pq.FormatTimestamp(time.Now())
	dateCreated, lastLogin := nowTime, nowTime

	queryStr = `
        INSERT INTO ` +
		conciergedb.ConciergeTables.Users + `
          (username, email, email_verified, date_created, last_login, password)
        VALUES ($1, $2, $3, $4, $5, $6)
        `

	Logger.Debug(
		"Checking query string for /access/signup",
		zap.String("query string", queryStr),
		zap.String("username", signupBody.User),
		zap.String("email", signupBody.Email),
		zap.Bool("email verified", emailVerified),
		zap.String("date created", string(dateCreated)),
		zap.String("last login", string(lastLogin)),
		zap.String("password hash", string(passwordHash)),
	)

	_, err = db.Query(
		queryStr,
		signupBody.User,
		signupBody.Email,
		emailVerified,
		dateCreated,
		lastLogin,
		passwordHash,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Error signing up new user"})
		return
	}

	var uid, gid, rid int
	uidErrorChan := make(chan error)
	gidErrorChan := make(chan error)
	ridErrorChan := make(chan error)
	db = GetDb()

	defer func() {
		close(uidErrorChan)
		close(gidErrorChan)
		close(ridErrorChan)
	}()

	go conciergedb.GetUid(signupBody.User, db, uidErrorChan, &uid)
	go conciergedb.GetGid(conciergedb.InitConciergeGroups.Site, db, gidErrorChan, &gid)
	go conciergedb.GetRid(conciergedb.InitConciergeRoles.User, db, ridErrorChan, &rid)
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

	queryStr = `
        INSERT INTO ` +
		conciergedb.ConciergeTables.GroupUsers + `
          (uid, gid)
        VALUES
          ($1, $2)
        `
	_, err = db.Query(
		queryStr,
		uid,
		gid,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Error setting up new user"})
		return
	}

	queryStr = `
        INSERT INTO ` +
		conciergedb.ConciergeTables.GroupUserRoles + `
          (uid, gid, rid)
        VALUES
          ($1, $2, $3)
        `
	_, err = db.Query(
		queryStr,
		uid,
		gid,
		rid,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Error setting up new user"})
		return
	}

	c.String(200, "Signed up successfully")
}

func Signin(c *gin.Context) {
	var signinBody SigninBody
	var err error
	var queryStr string
	var res *sql.Rows
	var passwordHash string
	var uid int
	db = GetDb()

	if err = c.ShouldBindJSON(&signinBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queryStr = `
        SELECT u.uid, u.password
        FROM ` +
		conciergedb.ConciergeTables.Users + ` u
        WHERE u.username = $1
        `

	Logger.Debug(
		"Checking query string for /access/signin",
		zap.String("query string", queryStr),
		zap.String("username", signinBody.User),
		zap.String("password", signinBody.Password),
	)

	res, err = db.Query(queryStr, signinBody.User)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Error finding user"})
		return
	} else if res.Next() {
		if err = res.Scan(&uid, &passwordHash); err != nil {
			Logger.Debug(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Error finding user"})
			return
		}
	} else {
		errString := fmt.Sprintf("User %s not found", signinBody.User)
		c.JSON(http.StatusNotFound, gin.H{"status": errString})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(signinBody.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "Password invalid"})
		return
	}

	claims := &ConciergeTokenClaims{
		signinBody.User,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + 86400,
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	var signedString string
	signedString, err = token.SignedString(GetJwtSecret())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Error authenticating user"})
	}

	lastLogin := pq.FormatTimestamp(time.Now())

	queryStr = `
        UPDATE ` +
		conciergedb.ConciergeTables.Users + `
        SET last_login = $1
        WHERE username = $2
        `
	_, err = db.Query(queryStr, lastLogin, signinBody.User)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Error updating user info"})
		return
	}

	signinRes := SigninRes{
		User:        signinBody.User,
		Auth:        true,
		AccessToken: signedString,
	}
	c.SecureJSON(http.StatusOK, signinRes)
}

func Signout(c *gin.Context) {
	c.String(http.StatusOK, "Signed out successfully")
}
