package minio

import (
	"encoding/xml"
	"errors"
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

func GetSuccessChunks(ctx *gin.Context) {
	var res int
	var uuid, uploaded, uploadID, chunks string
	fileMD5 := ctx.Query("md5")
	for {
		fileChunk, err := models.GetFileChunkByMD5(fileMD5)
		if err != nil {
			res = -1
			logger.LOG.Error("GetFileChunkByMD5 failed:", err.Error())
			break
		}

		if fileChunk == nil {
			break
		}

		uuid = fileChunk.UUID
		uploaded = strconv.Itoa(fileChunk.IsUploaded)
		uploadID = fileChunk.UploadID
		chunks = fileChunk.CompletedParts
	}

	ctx.JSON(200, map[string]string{
		"resultCode" : strconv.Itoa(res),
		"uuid": uuid,
		"uploaded": uploaded,
		"uploadID": uploadID,
		"chunks": chunks,
	})
}

func NewMultipart(ctx *gin.Context) {
	var res int
	var uuid, uploadID string

	for {
		totalChunkCounts := ctx.GetInt("totalChunkCounts")
		if totalChunkCounts > minio_ext.MaxPartsCount {
			res = -1
			logger.LOG.Errorf("totalChunkCounts(%d) is bigger than MaxPartsCount:", totalChunkCounts)
			break
		}

		fileSize := ctx.GetInt64("size")
		if fileSize > minio_ext.MaxMultipartPutObjectSize {
			res = -1
			logger.LOG.Errorf("fileSize(%u) is bigger than MaxMultipartPutObjectSize:", fileSize)
			break
		}

		uuid := gouuid.NewV4().String()
		uploadID, err := newMultiPartUpload(uuid)
		if err != nil {
			res = -1
			logger.LOG.Errorf("newMultiPartUpload failed:", err.Error())
			break
		}

		_, err = models.InsertFileChunk(&models.FileChunk{
			UUID:       uuid,
			UploadID:   uploadID,
			Md5:  		ctx.Query("md5"),
			Size:		fileSize,
			TotalChunks:totalChunkCounts,
		})

		if err != nil {
			ctx.Error(errors.New("500"))
			break
		}
	}

	ctx.JSON(200, map[string]string{
		"resultCode" : strconv.Itoa(res),
		"uuid": uuid,
		"uploadID":  uploadID,
	})
}

func GetMultipartUploadUrl(ctx *gin.Context) {
	var res int
	var url string
	uuid := ctx.Query("uuid")
	uploadID := ctx.Query("uploadID")
	partNumber := ctx.GetInt("chunkNumber")
	size := ctx.GetInt64("size")

	for {
		if size > minio_ext.MinPartSize {
			res = -1
			logger.LOG.Errorf("chunk size(%u) is too big", size)
			break
		}

		url,err := genMultiPartSignedUrl(uuid, uploadID, partNumber, size)
		if err != nil {
			res = -1
			logger.LOG.Error("genMultiPartSignedUrl failed:", err.Error())
			break
		}

		logger.LOG.Info(url)
	}

	ctx.JSON(200, map[string]string{
		"resultCode" : strconv.Itoa(res),
		"url": url,
	})
}

func CompleteMultipart(ctx *gin.Context) {
	var res int
	uuid := ctx.Query("uuid")
	uploadID := ctx.Query("uploadID")

	for {
		fileChunk, err := models.GetFileChunkByUUID(uuid)
		if err != nil {
			res = -1
			logger.LOG.Error("GetFileChunkByUUID failed:", err.Error())
			break
		}

		if fileChunk == nil {
			res = -1
			logger.LOG.Errorf("the record in file_chunk is not found, uuid(%s)", uuid)
			break
		}

		_, err = completeMultiPartUpload(uuid, uploadID, fileChunk.CompletedParts)
		if err != nil {
			res = -1
			logger.LOG.Error("completeMultiPartUpload failed:", err.Error())
			break
		}

		fileChunk.IsUploaded = models.FileUploaded

		err = models.UpdateFileChunk(fileChunk)
		if err != nil {
			res = -1
			logger.LOG.Error("UpdateFileChunk failed:", err.Error())
			break
		}
	}

	ctx.JSON(200, map[string]string{
		"resultCode" : strconv.Itoa(res),
	})
}

func UpdateMultipart(ctx *gin.Context) {
	var res int

	uuid := ctx.Query("uuid")
	partNumber := ctx.GetInt("chunkNumber")
	etag := ctx.Query("etag")

	for {
		fileChunk, err := models.GetFileChunkByUUID(uuid)
		if err != nil {
			res = -1
			logger.LOG.Error("GetFileChunkByUUID failed:", err.Error())
			break
		}

		if fileChunk == nil {
			res = -1
			logger.LOG.Errorf("the record in file_chunk is not found, uuid(%s)", uuid)
			break
		}

		fileChunk.CompletedParts += strconv.Itoa(partNumber) + "-" + strings.Replace(etag, "\"","", -1) + ","

		err = models.UpdateFileChunk(fileChunk)
		if err != nil {
			res = -1
			logger.LOG.Error("UpdateFileChunk failed:", err.Error())
			break
		}
	}

	ctx.JSON(200, map[string]string{
		"resultCode" : strconv.Itoa(res),
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