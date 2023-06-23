package controllers

import (
	"bipgit/internal/libgit"
	"bipgit/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Branch Attribution APIs
// @Summary 	Get Branch All Attributions
// @Description
// @Tags		Atrribution APIs
// @Security 	bearerAuth
// @Accept 		json
// @Produce 	json
// @Param 		branchName 				path 		string		 		true "Branch Name"
// @Param 		lastSyncedCommitId 		query 		string		 		true "Branch Name"
// @Success 	200 		{object} 	models.ApiResponse
// @Failure 	401 		{object} 	models.ApiResponse
// @Router 		/attribution/all/branch/{branchName} [get]
func GetBranchAllAttributions(c *gin.Context) {

	branchName, _ := c.Params.Get("branchName")
	lastSyncedCommitID := c.Query("lastSyncedCommitId")

	repo := middlewares.GetGitRepo(c)

	if branchName != "" {
		attributions, startCommitID, err := libgit.GetBranchAllAttributions(repo, branchName, lastSyncedCommitID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "There was some problem",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"data":          attributions,
			"startCommitId": startCommitID,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"error":   "Invalid Branch name",
	})
}

// Branch Attribution APIs
// @Summary 	Get Branch Attribution
// @Description
// @Tags		Atrribution APIs
// @Security 	bearerAuth
// @Accept 		json
// @Produce 	json
// @Param 		branchName 				path 		string		 		true "Branch Name"
// @Success 	200 		{object} 	models.ApiResponse
// @Failure 	401 		{object} 	models.ApiResponse
// @Router 		/attribution/branch/{branchName} [get]
func GetBranchAttributions(c *gin.Context) {

	branchName, _ := c.Params.Get("branchName")

	repo := middlewares.GetGitRepo(c)

	if branchName != "" {
		blockAttrs, err := libgit.GetBranchAttributions(repo, branchName)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "There was some problem",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    blockAttrs,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"error":   "Invalid Branch name",
	})
}
