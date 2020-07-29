package models

import (
	"oss/lib/mysql"

	"github.com/jinzhu/gorm"
)

const (
	FileNotUploaded int = iota
	FileUploaded
)

type FileChunk struct {
	gorm.Model
	UUID          string `gorm:"uuid UNIQUE"`
	Md5			  string `gorm:"INDEX"`
	IsUploaded    int `gorm:"DEFAULT 0"`  // not uploaded: 0, uploaded: 1
	UploadID   	  string	`gorm:"UNIQUE"`//minio upload id
	TotalChunks   int
	Size		  int64
	UserID        int64              `gorm:"INDEX"`
	CompletedParts		  []string	`gorm:"DEFAULT """`// chunkNumber+etag eg: ,1-asqwewqe21312312.2-123hjkas
}

// GetFileChunkByMD5 returns fileChunk by given md5
func GetFileChunkByMD5(md5 string) (*FileChunk, error) {
	fileChunk := new(FileChunk)
	mysql.Global.DB.Where("md5 = ?", md5).Find(&fileChunk)
	return fileChunk, nil
}

// GetFileChunkByMD5 returns fileChunk by given id
func GetFileChunkByMD5AndUser(md5 string, userID int64) (*FileChunk, error) {
	fileChunk := new(FileChunk)
	mysql.Global.DB.Where("md5 = ? and user_id = ?", md5, userID).Find(fileChunk)
	return fileChunk, nil
}

// GetAttachmentByID returns attachment by given id
func GetFileChunkByUUID(uuid string) (*FileChunk, error) {
	fileChunk := new(FileChunk)
	mysql.Global.DB.Where("uuid = ?", uuid).Find(fileChunk)
	return fileChunk, nil
}

// InsertFileChunk insert a record into file_chunk.
func InsertFileChunk(fileChunk *FileChunk) (_ *FileChunk, err error) {
	mysql.Global.DB.NewRecord(fileChunk)
	return fileChunk,nil
}

// UpdateFileChunk updates the given fileChunk in database
func UpdateFileChunk(fileChunk *FileChunk) error {
	mysql.Global.DB.Model(&fileChunk).Where("uuid = ?", fileChunk.UUID).Update("is_uploaded", "completed_parts")
	return nil
}
