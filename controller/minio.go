package controller

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	minio_service "oss/service/minio"
)


type Param struct {
	Username string `json:"username"`
	Password string `json:"password"`
	SourceObject string `json:"source_object"`
	Cmd string `json:"cmd"`
}

func MinioPost(c *gin.Context)  {
	var param Param
	success := false
	message := "falied"

	err := c.BindJSON(&param)
	if err != nil {
		message = "Failed to parse the req data:" + err.Error()
	} else {
		log.Println("cmd:", param.Cmd)
		switch param.Cmd {
		case "public_data":
			_,success,message = minio_service.PublicDataForUser(param.Username,param.SourceObject)
		case "init_user":
			_,success,message = minio_service.InitMinioForUser(param.Username,param.Password)
		default:
			message = "cmd error!cmd:" + param.Cmd
		}
	}

	log.Printf("success:%t, message:%s\n", success, message)

	c.JSON(200,gin.H{
		"success":success,
		"message":message,
	})
}

func MinioGet(c *gin.Context) {
	success := false
	message := "falied"
	objectInfo := ""
	cmd := c.Query("cmd")
	lastMarker := ""

	if cmd == "" {
		message = "Failed, req is err, no cmd"
	} else {
		log.Println("cmd:", cmd)
		switch cmd {
		case "list_objects":
			_,success,message = minio_service.ListSpecifiedObjects(c, &objectInfo, &lastMarker)
		default:
			message = "cmd error!cmd:" + cmd
		}
	}

	log.Printf("success:%t, message:%s\n", success, message)

	c.JSON(200,gin.H{
		"success":success,
		"message":message,
		"object_info" : objectInfo,
		"last_marker":lastMarker,
	})
}
