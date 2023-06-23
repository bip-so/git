package middlewares

import (
	"bipgit/internal/configs"
	"bipgit/internal/libgit"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	git "github.com/libgit2/git2go/v33"
)

const (
	LIBGIT_REPO = "LIBGIT_REPO"
	LIBGIT_SIGN = "LIBGIT_SIGN"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string
		auth := c.GetHeader("Authorization")
		if auth != "" && strings.HasPrefix(auth, "Bearer ") {
			tokenString = strings.ReplaceAll(auth, "Bearer ", "")
		}
		if tokenString == configs.GetSecretKey() {
			c.Next()
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
	}
}

func RepoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		studioId := c.GetHeader("x-studio-id")
		pageId := c.GetHeader("x-page-id")
		userName := c.GetHeader("x-user-name")
		userEmail := c.GetHeader("x-user-email")
		repoPath := path.Join(studioId, pageId)
		fmt.Println("API: ", c.Request.Method, c.Request.URL)
		fmt.Println("Repo: ", userName, userEmail, pageId, studioId)
		repo, sign := libgit.InitRepo(repoPath, userName, userEmail)
		defer repo.Repo.Free()
		defer repo.CloseConn()
		c.Set(LIBGIT_REPO, repo.Repo)
		c.Set(LIBGIT_SIGN, sign)
		c.Next()
	}
}

func GetGitRepo(c *gin.Context) *git.Repository {
	repo := c.MustGet(LIBGIT_REPO).(*git.Repository)
	return repo
}

func GetGitSign(c *gin.Context) *git.Signature {
	sign := c.MustGet(LIBGIT_SIGN).(*git.Signature)
	return sign
}
