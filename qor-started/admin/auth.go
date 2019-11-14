package admin

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	// "github.com/nerney/dappy"

	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/sirupsen/logrus"

	"qor-started/admin/ldap"
)

// Auth is a structure to handle authentication for QOR. It will satisify the
// qor.Auth interface.
type auth struct {
	session sessionConfig
	paths   pathConfig
}

type sessionConfig struct {
	name  string
	key   string
	store cookie.Store
}

type pathConfig struct {
	login  string
	logout string
	admin  string
}

type adminUser struct {
	Email     string `gorm:"not null;unique"`
	Brid      string `gorm:"not null;unique"`
	FirstName string
	LastName  string
	Password  []byte
	LastLogin *time.Time
}

func (u adminUser) DisplayName() string {
	if u.FirstName != "" && u.LastName != "" {
		return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
	}
	return u.Email
}

// GetLogin simply returns the login page
func (a *auth) GetLogin(c *gin.Context) {
	if sessions.Default(c).Get(a.session.key) != nil {
		fmt.Println("redirect happen!")
		c.Redirect(http.StatusSeeOther, a.paths.admin)
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

// PostLogin is the handler to check if the user can connect
func (a *auth) PostLogin(c *gin.Context) {
	session := sessions.Default(c)
	email := c.PostForm("email")
	password := c.PostForm("password")
	if email == "" || password == "" {
		c.Redirect(http.StatusSeeOther, a.paths.login)
		return
	}

	var client ldap.Client
	var err error

	// create a new client
	if client, err = ldap.New(ldap.Config{
		BaseDN: "dc=example,dc=com",
		Filter: "uid",
		ROUser: ldap.User{Name: "cn=read-only-admin,dc=example,dc=com", Pass: "password"},
		Host:   "ldap.forumsys.com:389",
	}); err != nil {
		panic(err)
	}

	// email and password to authenticate
	// email := "tesla"
	// password := "password"

	// attempt the authentication
	if err := client.Auth(email, password); err != nil {
		panic(err)
	} else {
		log.Println("Success!")
	}

	session.Set(a.session.key, email)
	fmt.Println(session)
	err = session.Save()
	if err != nil {
		logrus.WithError(err).Warn("Couldn't save session")
		c.Redirect(http.StatusSeeOther, a.paths.login)
		return
	}
	c.Redirect(http.StatusSeeOther, a.paths.admin)
}

// GetLogout allows the user to disconnect
func (a *auth) GetLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(a.session.key)
	if err := session.Save(); err != nil {
		logrus.WithError(err).Warn("Couldn't save session")
	}
	c.Redirect(http.StatusSeeOther, a.paths.login)
}

// GetCurrentUser satisfies the Auth interface and returns the current user
func (a auth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
	// var userid uint
	var email string

	s, err := a.session.store.Get(c.Request, a.session.name)
	if err != nil {
		return nil
	}
	if v, ok := s.Values[a.session.key]; ok {
		email = v.(string)
		fmt.Println(email)
	} else {
		return nil
	}

	var AdminUser adminUser
	AdminUser.Email = "Administrator"
	return AdminUser

	// return nil
}

// LoginURL statisfies the Auth interface and returns the route used to log
// users in
func (a auth) LoginURL(c *admin.Context) string { // nolint: unparam
	return a.paths.login
}

// LogoutURL statisfies the Auth interface and returns the route used to logout
// a user
func (a auth) LogoutURL(c *admin.Context) string { // nolint: unparam
	return a.paths.logout
}
