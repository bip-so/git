package controllers

import (
	"bipgit/internal/libgit"
	"bipgit/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Create Branches
// @Summary		Create Branches API
// @Description
// @Tags		Branch APIs
// @Security 	bearerAuth
// @Accept 		json
// @Produce 	json
// @Param 		branchName 	path 		string		 		true "Branch Name"
// @Success 	200 		{object} 	models.ApiResponse
// @Failure 	401 		{object} 	models.ApiResponse
// @Router 		/branch/create/{branchName} [post]
func CreateBranch(c *gin.Context) {

	fromBranchName, _ := c.Params.Get("branchName")

	var bodyData struct {
		BranchName string `json:"branchName"`
	}
	c.BindJSON(&bodyData)

	repo := middlewares.GetGitRepo(c)
	// sign := middlewares.GetGitSign(c)

	if bodyData.BranchName != "" {
		commitID, err := libgit.CreateNewBranch(repo, fromBranchName, bodyData.BranchName)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "There was some problem",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    commitID,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"error":   "Invalid Branch name",
	})
}

// Delete Branch
// @Summary		Delete Branch API
// @Description
// @Tags		Branch APIs
// @Security 	bearerAuth
// @Accept 		json
// @Produce 	json
// @Param 		branchName 	path 		string		 		true "Branch Name"
// @Success 	200 		{object} 	models.ApiResponse
// @Failure 	401 		{object} 	models.ApiResponse
// @Router 		/branch/{branchName} [delete]
func DeleteBranch(c *gin.Context) {

	branchName, _ := c.Params.Get("branchName")

	repo := middlewares.GetGitRepo(c)
	// sign := middlewares.GetGitSign(c)

	if branchName != "" {
		err := libgit.DeleteBranch(repo, branchName)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "There was some problem",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"error":   "Invalid Branch name",
	})
}
