package server

import (
	"bipgit/docs"
	"bipgit/internal/controllers"
	"bipgit/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitServer() {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	/* ---------------------------  Public Swagger routes  --------------------------- */
	api := router.Group("/")
	docs.SwaggerInfo.BasePath = "/api"
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	/* ---------------------------  bip git backend routes  --------------------------- */
	api.GET("/health", health)
	router.Use(middlewares.AuthMiddleware())
	router.Use(middlewares.RepoMiddleware())
	api = router.Group("/api")
	{
		// Branch
		branchGroup := api.Group("branch")
		{
			branchGroup.POST("/create/:branchName", controllers.CreateBranch)
			branchGroup.DELETE("/:branchName", controllers.DeleteBranch)
		}

		// Snapshots
		snapshotGroup := api.Group("snapshot")
		{
			snapshotGroup.POST("/create", controllers.CreateSnapshot)
			snapshotGroup.GET("/branch/:branchName", controllers.GetBranchSnapshot)
			snapshotGroup.GET("/get/:commitId", controllers.GetSnapshotById)
			snapshotGroup.POST("/block/create", controllers.CreateBlockSnapshot)
		}

		// Merge Request
		mergeGroup := api.Group("mergereq")
		{
			mergeGroup.POST("/merge", controllers.MergeBranches)
		}

		// History
		logGroup := api.Group("history")
		{
			logGroup.GET("/branch/:branchName", controllers.GetBranchHistory)
		}

		// Attribution
		attributionGroup := api.Group("attribution")
		{
			attributionGroup.GET("/all/branch/:branchName", controllers.GetBranchAllAttributions)
			attributionGroup.GET("/branch/:branchName", controllers.GetBranchAttributions)
		}
	}

	router.Run(":9004")
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
