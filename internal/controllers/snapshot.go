package controllers

import (
	"bipgit/internal/bip"
	"bipgit/internal/constants"
	"bipgit/internal/libgit"
	"bipgit/internal/middlewares"
	"bipgit/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Create Snapshot
// @Summary		Create snapshot
// @Description
// @Tags		Snapshot APIs
// @Security 	bearerAuth
// @Accept 		json
// @Produce 	json
// @Param 		body 		body 		models.CreateSnapshotData true "Create Snapshot Data"
// @Success 	200 		{object} 	models.ApiResponse
// @Failure 	401 		{object} 	models.ApiResponse
// @Router 		/snapshot/create [post]
func CreateSnapshot(c *gin.Context) {
	var bodyData struct {
		Blocks         models.BEBlocksSlice `json:"blocks"`
		BranchName     string               `json:"branchName"`
		FromBranchName string               `json:"fromBranchName"`
		Message        string               `json:"message"`
	}
	err := c.ShouldBindJSON(&bodyData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "invalid body",
		})
		return
	}

	repo := middlewares.GetGitRepo(c)
	sign := middlewares.GetGitSign(c)

	branchExists, err := libgit.CheckIfBranchExists(repo, bodyData.BranchName)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "There was some problem",
		})
		return
	}
	if !branchExists && bodyData.BranchName != constants.DefaultBranchName {
		_, err := libgit.CreateNewBranch(repo, bodyData.FromBranchName, bodyData.BranchName)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "There was some problem",
			})
			return
		}
	}

	commitId, err := libgit.WriteBlocksAsCommit(repo, sign, bodyData.Blocks.ConvertToBlocksFromBEBlocks(), bodyData.BranchName, bodyData.Message)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "There was some problem",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    commitId,
	})
}

// Get Branch Snapshot
// @Summary 	Get Branch snapshot
// @Description
// @Tags		Snapshot APIs
// @Security 	bearerAuth
// @Accept 		json
// @Produce 	json
// @Param 		branchName 	path 		string		 		true "Branch Name"
// @Success 	200 		{object} 	models.ApiResponse
// @Failure 	401 		{object} 	models.ApiResponse
// @Router 		/snapshot/branch/{branchName} [get]
func GetBranchSnapshot(c *gin.Context) {

	branchName, _ := c.Params.Get("branchName")

	repo := middlewares.GetGitRepo(c)
	sign := middlewares.GetGitSign(c)

	blocks, err := libgit.ReadAllBlocksFromBranch(repo, sign, branchName)
	if err != nil {
		if err.Error() == "invalid branch" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Invalid branch",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "There was some problem",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    blocks.ConvertToBEBlocksFromBlocks(),
	})
}

// Get Snapshot By ID
// @Summary 	Get Snapshot By ID
// @Description
// @Tags		Snapshot APIs
// @Security 	bearerAuth
// @Accept 		json
// @Produce 	json
// @Param 		commitId 	path 		string		 		true "Branch Name"
// @Success 	200 		{object} 	models.ApiResponse
// @Failure 	401 		{object} 	models.ApiResponse
// @Router 		/snapshot/get/{commitId} [get]
func GetSnapshotById(c *gin.Context) {

	commitId, _ := c.Params.Get("commitId")

	repo := middlewares.GetGitRepo(c)
	sign := middlewares.GetGitSign(c)

	if commitId != "" {
		blocks, err := libgit.ReadAllBlocksFromCommit(repo, sign, commitId)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "There was some problem",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    blocks.ConvertToBEBlocksFromBlocks(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"error":   "Invalid Commit",
	})
}

// Create Block Snapshot
// @Summary		Create Block snapshot
// @Description
// @Tags		Snapshot APIs
// @Security 	bearerAuth
// @Accept 		json
// @Produce 	json
// @Param 		body 		body 		models.CreateBlockSnapshotData true "Create Snapshot Data"
// @Success 	200 		{object} 	models.ApiResponse
// @Failure 	401 		{object} 	models.ApiResponse
// @Router 		/snapshot/block/create [post]
func CreateBlockSnapshot(c *gin.Context) {
	var bodyData struct {
		Blocks         models.BEBlocksSlice `json:"blocks"`
		BranchName     string               `json:"branchName"`
		MessageBlockID string               `json:"messageBlockId"`
	}
	c.BindJSON(&bodyData)

	repo := middlewares.GetGitRepo(c)
	sign := middlewares.GetGitSign(c)

	branchExists, err := libgit.CheckIfBranchExists(repo, bodyData.BranchName)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "There was some problem",
		})
		return
	}
	if !branchExists && bodyData.BranchName != constants.DefaultBranchName {
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"error":   "Branch does not exist",
			})
			return
		}
	}

	commitId, err := bip.CommitNewMessageBlock(repo, sign, bodyData.Blocks.ConvertToBlocksFromBEBlocks(), bodyData.BranchName, bodyData.MessageBlockID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "CommitNewMessageBlock: There was some problem",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    commitId,
	})
}
