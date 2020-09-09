package models

import (
	"github.com/jinzhu/gorm"

	"oss/lib/mysql"
)

const (
	FileNotUploaded int = iota
	FileUploaded
)

type FileChunk struct {
	//ID			  int64  `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	gorm.Model

	UUID          string `gorm:"UNIQUE"`
	Md5			  string `gorm:"UNIQUE"`
	IsUploaded    int `gorm:"DEFAULT 0"`  // not uploaded: 0, uploaded: 1
	UploadID   	  string	`gorm:"UNIQUE"`//minio upload id
	TotalChunks   int
	Size		  int64
	FileName	  string
	CompletedParts		  string	`gorm:"type:text"`// chunkNumber+etag eg: ,1-asqwewqe21312312.2-123hjkas
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
	if err := mysql.Global.DB.Where("md5 = ?", md5).Find(&fileChunk).Error; err != nil {
		return fileChunk, err
	}
	return fileChunk, nil
}

// GetFileChunkByUUID returns attachment by given uuid
func GetFileChunkByUUID(uuid string) (*FileChunk, error) {
	fileChunk := new(FileChunk)
	if err := mysql.Global.DB.Where("uuid = ?", uuid).Find(&fileChunk).Error; err != nil {
		return fileChunk, err
	}
	return fileChunk, nil
}

// InsertFileChunk insert a record into file_chunk.
func InsertFileChunk(fileChunk *FileChunk) (_ *FileChunk, err error) {
	if err := mysql.Global.DB.Create(fileChunk).Error; err != nil {
		return fileChunk, err
	}
	return fileChunk,nil
}

// UpdateFileChunk updates the given fileChunk in database
func UpdateFileChunk(fileChunk *FileChunk) error {
	if err := mysql.Global.DB.Model(&fileChunk).Where("uuid = ?", fileChunk.UUID).
		Updates(FileChunk{IsUploaded:fileChunk.IsUploaded, CompletedParts:fileChunk.CompletedParts}).Error; err != nil {
		return err
	}
	return nil
}
