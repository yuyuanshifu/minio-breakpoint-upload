package config


import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"os"

	logger "oss/lib/log"
	"oss/lib/rsa"

	"github.com/json-iterator/go"
)

var MysqlIp string
var MysqlUsername string
var MysqlPassword string
var MysqlPort string
var MysqlDbName string
var PORT string
var MinioAddress string
var MinioAccessKeyId string
var MinioSecretAccessKey string
var MinioSecure string
var MinioBucket string
var MinioBasePath string
var MinioLocation string


func loadFromConfigFile(configFilePath string)error{
	file,err:= os.Open(configFilePath)
	if err!= nil{
		logger.LOG.Error(err)
         return err
	}

	data,err := ioutil.ReadAll(file)

	if nil != err{
		return err
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var jsonConfig jsoniter.Any = json.Get(data)

	MysqlIp = jsonConfig.Get("MYSQL_IP").ToString()
	MysqlUsername = jsonConfig.Get("MYSQL_USERNAME").ToString()
	MysqlPassword = jsonConfig.Get("MYSQL_PASSWORD").ToString()
	MysqlDbName = jsonConfig.Get("MYSQL_DBNAME").ToString()
	MysqlPort = jsonConfig.Get("MYSQL_PORT").ToString()
	PORT = jsonConfig.Get("PORT").ToString()
	MinioAddress = jsonConfig.Get("MINIO_ADDRESS").ToString()
	MinioAccessKeyId = jsonConfig.Get("MINIO_ACCESS_KEY_ID").ToString()
	keyTmp := jsonConfig.Get("MINIO_SECRET_ACCESS_KEY").ToString()
	MinioSecure = jsonConfig.Get("MINIO_SECURE").ToString()
	MinioBucket = jsonConfig.Get("MINIO_BUCKET").ToString()
	MinioBasePath = jsonConfig.Get("MINIO_BASE_PATH").ToString()
	MinioLocation = jsonConfig.Get("MINIO_LOCATION").ToString()

	if MysqlIp == "" || MysqlUsername == "" || MysqlPassword == "" || MysqlPort == "" || PORT == "" || MysqlDbName == "" || MinioAddress == "" || MinioAccessKeyId == "" || keyTmp == "" || MinioSecure == "" {
		return errors.New("config is error")
	}

	enc,err := base64.StdEncoding.DecodeString(keyTmp)
	if err != nil {
		return err
	}

	dec,err := rsa.RsaDecrypt([]byte(enc))
	if err != nil {
		return err
	}

	MinioSecretAccessKey = string(dec)

	return nil
}

func init(){
	configFile := "config.json"
	err:= loadFromConfigFile(configFile)
	if nil != err{
		logger.LOG.Fatal("Failed to load config,Error:" + err.Error())
		return
	}

	return
}
