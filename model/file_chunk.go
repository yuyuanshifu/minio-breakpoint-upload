package models

import (
	"github.com/pkg/errors"
	"oss/lib/mysql"

	"github.com/jinzhu/gorm"
)

const (
	FileNotUploaded int = iota
	FileUploaded
)

type FileChunk struct {
	//ID			  int64  `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	gorm.Model
	UUID          string `gorm:"UNIQUE"`
	Md5			  string `gorm:"INDEX"`
	IsUploaded    int `gorm:"DEFAULT 0"`  // not uploaded: 0, uploaded: 1
	UploadID   	  string	`gorm:"UNIQUE"`//minio upload id
	TotalChunks   int
	Size		  int64
	CompletedParts		  string	`gorm:"type:text,DEFAULT """`// chunkNumber+etag eg: ,1-asqwewqe21312312.2-123hjkas
}

func init() {
	if !mysql.Global.DB.HasTable(&FileChunk{}) {
		mysql.Global.DB.CreateTable(&FileChunk{})
	}
	mysql.Global.DB.AutoMigrate(&FileChunk{})
}

// GetFileChunkByMD5 returns fileChunk by given md5
func GetFileChunkByMD5(md5 string) (*FileChunk, error) {
	fileChunk := new(FileChunk)
	if !mysql.Global.DB.Where("md5 = ?", md5).Find(&fileChunk).RecordNotFound(){
		return fileChunk, errors.New("error")
	}
	return fileChunk, nil
}

// GetAttachmentByID returns attachment by given id
func GetFileChunkByUUID(uuid string) (*FileChunk, error) {
	fileChunk := new(FileChunk)
	if err := mysql.Global.DB.Where("uuid = ?", uuid).Find(&fileChunk).Error; err != nil {
		return fileChunk, err
	}
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
