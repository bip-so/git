package controllers

import (
	"bipgit/internal/constants"
	"bipgit/internal/libgit"
	"bipgit/internal/middlewares"
	"bipgit/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Merge Branches
// @Summary		Merge Branches API
// @Description
// @Tags		Merge Branches
// @Security 	bearerAuth
// @Accept 		json
// @Produce 	json
// @Param 		body 		body 		models.MergeBranchesData true "Merge Branches Data"
// @Success 	200 		{object} 	models.ApiResponse
// @Failure 	401 		{object} 	models.ApiResponse
// @Router 		/mergereq/merge [post]
func MergeBranches(c *gin.Context) {
	var bodyData struct {
		ToBranchName   string `json:"toBranchName"`
		FromBranchName string `json:"fromBranchName"`
		MergeType      string `json:"mergeType"`
		// FromBranchCreatedCommitID string                 `json:"fromBranchCreatedCommitId"`
		ChangesAccepted map[string]interface{} `json:"changesAccepted"`
	}
	c.BindJSON(&bodyData)

	repo := middlewares.GetGitRepo(c)
	sign := middlewares.GetGitSign(c)
	if !(utils.ExistsIn(bodyData.MergeType, []string{constants.MergeRequestAccepted, constants.MergeRequestPartiallyAccepted})) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "invalid mergetype",
		})
		return
	}

	commitID, srcCommitID, destCommitID, err := libgit.MergeBranches(repo, sign, bodyData.FromBranchName, bodyData.ToBranchName, bodyData.MergeType, bodyData.ChangesAccepted)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "There was some problem",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]string{
			"commitId":     commitID,
			"srcCommitId":  srcCommitID,
			"destCommitId": destCommitID,
		},
	})
}
