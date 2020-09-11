package main

import (
	"oss/config"
	_ "oss/docs"
	"oss/lib/cors"
	logger "oss/lib/log"
	minioService "oss/service/minio"

	"github.com/gin-gonic/gin"
	gs "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func  main()  {
	router := gin.New()
	router.Use(cors.Cors())

	router.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))

	minio := router.Group("/minio")
	{
		minio.GET("/get_chunks", minioService.GetSuccessChunks)
		minio.GET("/new_multipart", minioService.NewMultipart)
		minio.GET("/get_multipart_url", minioService.GetMultipartUploadUrl)
		minio.POST("/complete_multipart", minioService.CompleteMultipart)
		minio.POST("/update_chunk", minioService.UpdateMultipart)
	}

	router.Run(":" + config.PORT)

	logger.LOG.Infof("service is running on port:", config.PORT)

}