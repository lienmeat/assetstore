package main

import (
	"os"
	"strconv"

	"assetstore"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"
)

func main() {
	//default log level is info, but evn var can override
	log.SetLevel(log.InfoLevel)
	if lvl, err := log.ParseLevel(os.Getenv("LOG_LEVEL")); err != nil {
		log.SetLevel(lvl)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	dnm := assetstore.NewDynamoDBMetaTokenStore(os.Getenv("DYNAMODB_TABLE"), sess)

	//AssetStorage implements all the required interfaces required in one abstraction
	assetStorage := assetstore.NewAssetStorage(
		dnm,
		dnm,
		assetstore.NewS3Storage(os.Getenv("S3_BUCKET"), sess),
	)

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic("PORT env var was not correctly defined")
	}

	//run an http/api server to store/get assets
	assetstore.RunAPI(assetStorage, assetStorage, assetStorage, os.Getenv("BASE_PATH"), port)
}