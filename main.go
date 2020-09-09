package main

import (
	"oss/config"
	"oss/lib/cors"
	logger "oss/lib/log"
	minioService "oss/service/minio"

	"github.com/gin-gonic/gin"
)

func  main()  {
	router := gin.New()
	router.Use(cors.Cors())

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