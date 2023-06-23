package controllers

import (
	"bipgit/internal/libgit"
	"bipgit/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Get Branch History
// @Summary 	Get Branch History
// @Description
// @Tags		Branch History Log APIs
// @Security 	bearerAuth
// @Accept 		json
// @Produce 	json
// @Param 		branchName 	path 		string		 		true "Branch Name"
// @Param 		start 		query 		string		 		true "Branch Name"
// @Success 	200 		{object} 	models.ApiResponse
// @Failure 	401 		{object} 	models.ApiResponse
// @Router 		/history/branch/{branchName} [get]
func GetBranchHistory(c *gin.Context) {

	branchName, _ := c.Params.Get("branchName")
	startCommitID := c.Query("start")

	repo := middlewares.GetGitRepo(c)
	// sign := middlewares.GetGitSign(c)

	if branchName != "" {
		logs, next, err := libgit.GetLogsOfBranch(repo, branchName, startCommitID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "There was some problem",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    logs,
			"next":    next,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"error":   "Invalid Branch name",
	})
}
