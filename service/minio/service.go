package minio

import (
	"encoding/xml"
	"net/http"
	"oss/config"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	logger "oss/lib/log"
	"oss/lib/minio_ext"
	"oss/model"

	"github.com/gin-gonic/gin"
	miniov6 "github.com/minio/minio-go/v6"
	gouuid "github.com/satori/go.uuid"
)

const (
	PresignedUploadPartUrlExpireTime = time.Hour * 24 * 7
)

type ComplPart struct {
	PartNumber int	`json:"partNumber"`
	ETag string	`json:"eTag"`
}

type CompleteParts struct {
	Data []ComplPart	`json:"completedParts"`
}
// completedParts is a collection of parts sortable by their part numbers.
// used for sorting the uploaded parts before completing the multipart request.
type completedParts []miniov6.CompletePart
func (a completedParts) Len() int           { return len(a) }
func (a completedParts) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a completedParts) Less(i, j int) bool { return a[i].PartNumber < a[j].PartNumber }

// completeMultipartUpload container for completing multipart upload.
type completeMultipartUpload struct {
	XMLName xml.Name       `xml:"http://s3.amazonaws.com/doc/2006-03-01/ CompleteMultipartUpload" json:"-"`
	Parts   []miniov6.CompletePart `xml:"Part"`
}

func NewMultipart(ctx *gin.Context) {
	var uuid, uploadID string

	totalChunkCounts,err := strconv.Atoi(ctx.Query("totalChunkCounts"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "totalChunkCounts is illegal.")
		return
	}

	if totalChunkCounts > minio_ext.MaxPartsCount || totalChunkCounts <= 0{
		ctx.JSON(http.StatusBadRequest, "totalChunkCounts is illegal.")
		return
	}

	fileSize,err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "size is illegal.")
		return
	}

	if fileSize > minio_ext.MaxMultipartPutObjectSize || fileSize <= 0{
		ctx.JSON(http.StatusBadRequest, "size is illegal.")
		return
	}

	uuid = gouuid.NewV4().String()
	uploadID, err = newMultiPartUpload(uuid)
	if err != nil {
		logger.LOG.Errorf("newMultiPartUpload failed:", err.Error())
		ctx.JSON(http.StatusInternalServerError, "newMultiPartUpload failed.")
		return
	}

	_, err = models.InsertFileChunk(&models.FileChunk{
		UUID:       uuid,
		UploadID:   uploadID,
		Md5:  		ctx.Query("md5"),
		Size:		fileSize,
		FileName:   ctx.Query("fileName"),
		TotalChunks:totalChunkCounts,
	})

	if err != nil {
		logger.LOG.Error("InsertFileChunk failed:", err.Error())
		ctx.JSON(http.StatusInternalServerError, "InsertFileChunk failed.")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"uuid": uuid,
		"uploadID":  uploadID,
	})
}

func GetMultipartUploadUrl(ctx *gin.Context) {
	var url string
	uuid := ctx.Query("uuid")
	uploadID := ctx.Query("uploadID")

	partNumber,err := strconv.Atoi(ctx.Query("chunkNumber"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "chunkNumber is illegal.")
		return
	}

	size,err := strconv.ParseInt(ctx.Query("size"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "size is illegal.")
		return
	}
	if size > minio_ext.MinPartSize {
		ctx.JSON(http.StatusBadRequest, "size is illegal.")
		return
	}

	url,err = genMultiPartSignedUrl(uuid, uploadID, partNumber, size)
	if err != nil {
		logger.LOG.Error("genMultiPartSignedUrl failed:", err.Error())
		ctx.JSON(http.StatusInternalServerError, "genMultiPartSignedUrl failed.")
		return
	}


	ctx.JSON(http.StatusOK, gin.H {
		"url": url,
	})
}

func CompleteMultipart(ctx *gin.Context) {
	uuid := ctx.PostForm("uuid")
	uploadID := ctx.PostForm("uploadID")

	fileChunk, err := models.GetFileChunkByUUID(uuid)
	if err != nil {
		logger.LOG.Error("GetFileChunkByUUID failed:", err.Error())
		ctx.JSON(http.StatusInternalServerError, "GetFileChunkByUUID failed.")
		return
	}

	_, err = completeMultiPartUpload(uuid, uploadID, fileChunk.CompletedParts)
	if err != nil {
		logger.LOG.Error("completeMultiPartUpload failed:", err.Error())
		ctx.JSON(http.StatusInternalServerError, "completeMultiPartUpload failed.")
		return
	}

	fileChunk.IsUploaded = models.FileUploaded

	err = models.UpdateFileChunk(fileChunk)
	if err != nil {
		logger.LOG.Error("UpdateFileChunk failed:", err.Error())
		ctx.JSON(http.StatusInternalServerError, "UpdateFileChunk failed.")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
	})
}

func UpdateMultipart(ctx *gin.Context) {
	uuid := ctx.PostForm("uuid")
	etag := ctx.PostForm("etag")

	fileChunk, err := models.GetFileChunkByUUID(uuid)
	if err != nil {
		logger.LOG.Error("GetFileChunkByUUID failed:", err.Error())
		ctx.JSON(http.StatusInternalServerError, "GetFileChunkByUUID failed.")
		return
	}

	fileChunk.CompletedParts += ctx.PostForm("chunkNumber") + "-" + strings.Replace(etag, "\"","", -1) + ","

	err = models.UpdateFileChunk(fileChunk)
	if err != nil {
		logger.LOG.Error("UpdateFileChunk failed:", err.Error())
		ctx.JSON(http.StatusInternalServerError, "UpdateFileChunk failed.")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
	})
}

func newMultiPartUpload(uuid string) (string, error){
	_, core, _, err := getClients()
	if err != nil {
		logger.LOG.Error("getClients failed:", err.Error())
		return "", err
	}

	bucketName := config.MinioBucket
	objectName := strings.TrimPrefix(path.Join(config.MinioBasePath, path.Join(uuid[0:1], uuid[1:2], uuid)), "/")

	return core.NewMultipartUpload(bucketName, objectName, miniov6.PutObjectOptions{})
}

func genMultiPartSignedUrl(uuid string, uploadId string, partNumber int, partSize int64) (string, error) {
	_, _, minioClient, err := getClients()
	if err != nil {
		logger.LOG.Error("getClients failed:", err.Error())
		return "", err
	}

	bucketName := config.MinioBucket
	objectName := strings.TrimPrefix(path.Join(config.MinioBasePath, path.Join(uuid[0:1], uuid[1:2], uuid)), "/")

	return minioClient.GenUploadPartSignedUrl(uploadId, bucketName, objectName, partNumber, partSize, PresignedUploadPartUrlExpireTime, config.MinioLocation)

}

func completeMultiPartUpload(uuid string, uploadID string, complParts string) (string, error){
	_, core, _, err := getClients()
	if err != nil {
		logger.LOG.Error("getClients failed:", err.Error())
		return "", err
	}

	bucketName := config.MinioBucket
	objectName := strings.TrimPrefix(path.Join(config.MinioBasePath, path.Join(uuid[0:1], uuid[1:2], uuid)), "/")

	var complMultipartUpload completeMultipartUpload
	for _,part := range strings.Split(complParts, ",") {
		if part == "" {
			break
		}
		partNumber, err := strconv.Atoi(strings.Split(part,"-")[0])
		if err != nil {
			logger.LOG.Error(err.Error())
			return "",err
		}
		complMultipartUpload.Parts = append(complMultipartUpload.Parts, miniov6.CompletePart{
			PartNumber: partNumber,
			ETag: strings.Split(part,"-")[1],
		})
	}

	// Sort all completed parts.
	sort.Sort(completedParts(complMultipartUpload.Parts))

	return core.CompleteMultipartUpload(bucketName, objectName, uploadID, complMultipartUpload.Parts)
}

func GetSuccessChunks(ctx *gin.Context) {
	var res = -1
	var uuid, uploaded, uploadID, chunks string

	fileMD5 := ctx.Query("md5")
	for {
		fileChunk, err := models.GetFileChunkByMD5(fileMD5)
		if err != nil {
			logger.LOG.Error("GetFileChunkByMD5 failed:", err.Error())
			break
		}

		uuid = fileChunk.UUID
		uploaded = strconv.Itoa(fileChunk.IsUploaded)
		uploadID = fileChunk.UploadID

		bucketName := config.MinioBucket
		objectName := strings.TrimPrefix(path.Join(config.MinioBasePath, path.Join(uuid[0:1], uuid[1:2], uuid)), "/")

		isExist, err := isObjectExist(bucketName, objectName)
		if err != nil {
			logger.LOG.Error("isObjectExist failed:", err.Error())
			break
		}

		if isExist {
			uploaded = "1"
			if fileChunk.IsUploaded != models.FileUploaded {
				logger.LOG.Info("the file has been uploaded but not recorded")
				fileChunk.IsUploaded = 1
				if err = models.UpdateFileChunk(fileChunk); err != nil {
					logger.LOG.Error("UpdateFileChunk failed:", err.Error())
				}
			}
			res = 0
			break
		} else {
			uploaded = "0"
			if fileChunk.IsUploaded == models.FileUploaded {
				logger.LOG.Info("the file has been recorded but not uploaded")
				fileChunk.IsUploaded = 0
				if err = models.UpdateFileChunk(fileChunk); err != nil {
					logger.LOG.Error("UpdateFileChunk failed:", err.Error())
				}
			}
		}

		_, _, client, err := getClients()
		if err != nil {
			logger.LOG.Error("getClients failed:", err.Error())
			break
		}

		partInfos, err := client.ListObjectParts(bucketName, objectName, uploadID)
		if err != nil {
			logger.LOG.Error("ListObjectParts failed:", err.Error())
			break
		}

		for _, partInfo := range partInfos {
			chunks += strconv.Itoa(partInfo.PartNumber) + "-" + partInfo.ETag + ","
		}

		break
	}

	logger.LOG.Info(chunks)

	ctx.JSON(http.StatusOK, gin.H {
		"resultCode" : strconv.Itoa(res),
		"uuid": uuid,
		"uploaded": uploaded,
		"uploadID": uploadID,
		"chunks": chunks,
	})
}

func isObjectExist(bucketName string, objectName string) (bool, error) {
	isExist := false
	doneCh := make(chan struct{})
	defer close(doneCh)

	client, _, _, err := getClients()
	if err != nil {
		logger.LOG.Error("getClients failed:", err.Error())
		return isExist, err
	}

	objectCh := client.ListObjects(bucketName, objectName, false, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			logger.LOG.Error(object.Err)
			return isExist, object.Err
		}
		isExist = true
		break
	}

	return isExist, nil
}