package mysql

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"oss/config"
	logger "oss/lib/log"
)

type global struct {
	DB *gorm.DB
}

var Global global

func init() {
	//enc,err := base64.StdEncoding.DecodeString(config.MysqlPassword)
	//if err != nil {
	//	logger.LOG.Error("DecodeString failed:", err.Error())
	//	return
	//}
	//
	//dec,err := rsa.RsaDecrypt([]byte(enc))
	//if err != nil {
	//	logger.LOG.Error("RsaDecrypt failed:", err.Error())
	//	return
	//}

	dec := config.MysqlPassword

	dbDriver := config.MysqlUsername + ":" + string(dec) + "@tcp(" + config.MysqlIp + ":" + config.MysqlPort + ")/" + config.MysqlDbName + "?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open("mysql", dbDriver)
	if err != nil {
		logger.LOG.Error("open db failed:" + err.Error())
		return
	}

	//defer db.Close()

	db.SingularTable(true)
	db.LogMode(true)

	// SetMaxIdleCons 设置连接池中的最大闲置连接数。
	db.DB().SetMaxIdleConns(10)

	// SetMaxOpenCons 设置数据库的最大连接数量。
	db.DB().SetMaxOpenConns(100)

	Global.DB = db

	return
}
