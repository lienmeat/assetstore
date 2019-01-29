package assetstore

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	log "github.com/sirupsen/logrus"
)

//many adapters/backends could exist for storing asset meta and tokens
//dynamodb will do nicely

const (
	ASSET_KEY_PREFIX = "ASSET_"
	TOKEN_KEY_PREFIX = "TOKEN_"
)

type DynamoDBMetaTokenStore struct {
	table string
	*dynamodb.DynamoDB
}

func NewDynamoDBMetaTokenStore(table string, sess *session.Session) *DynamoDBMetaTokenStore {
	return &DynamoDBMetaTokenStore{
		table: table,
		DynamoDB: dynamodb.New(sess),
	}
}

func (s *DynamoDBMetaTokenStore) GetMeta(id string) (meta AssetMeta, err error) {
	if id == "" {
		return meta, fmt.Errorf("zero-length id")
	}
	result, err := s.Query(&dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": {
				S: aws.String(ASSET_KEY_PREFIX + id),
			},
		},
		KeyConditionExpression: aws.String("ObjID = :v1"),
		TableName: aws.String(s.table),
	})
	if err != nil {
		return
	}
	if *result.Count != int64(1) {
		return meta, fmt.Errorf("could not find result for asset with id %s", id)
	}
	obj := result.Items[0]
	return dynamoAssetAttrMapToMeta(obj), err
}

func (s *DynamoDBMetaTokenStore) StoreMeta(meta AssetMeta) (err error) {
	if !meta.Valid() {
		return fmt.Errorf("meta invalid")
	}
	_, err = s.PutItem(&dynamodb.PutItemInput{
		Item:      assetMetaToDynamoAttrMap(meta),
		TableName: aws.String(s.table),
	})
	return
}

func (s *DynamoDBMetaTokenStore) GetToken(token string) (t AssetToken, err error) {
	if token == "" {
		return t, fmt.Errorf("zero-length token")
	}
	result, err := s.Query(&dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": {
				S: aws.String(TOKEN_KEY_PREFIX + token),
			},
		},
		KeyConditionExpression: aws.String("ObjID = :v1"),
		TableName: aws.String(s.table),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"context": "DynamoDBMetaTokenStore.GetToken()",
			"token": token,
			"table": s.table,
			"result": result,

		}).Error(err)
		return
	}
	if *result.Count != int64(1) {
		return t, fmt.Errorf("could not find result for token %s", token)
	}
	obj := result.Items[0]
	t = dynamoTokenAttrMapToAssetToken(obj)
	if !time.Now().Before(time.Unix(t.Expiry, 0)) {
		return t, fmt.Errorf("token expired")
	}
	return
}

func (s *DynamoDBMetaTokenStore) StoreToken(token AssetToken) (err error) {
	if !token.Valid() {
		return fmt.Errorf("token invalid")
	}
	_, err = s.PutItem(&dynamodb.PutItemInput{
		Item:      assetTokenToDynamoAttrMap(token),
		TableName: aws.String(s.table),
	})
	return
}

func assetMetaToDynamoAttrMap(meta AssetMeta) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		"ObjID": {
			S:  aws.String(ASSET_KEY_PREFIX + meta.ID),
		},
		"ObjSort": {
			S: aws.String(strconv.Itoa(meta.Version)),
		},
		"AssetName": {
			S: aws.String(meta.Name),
		},
		"Size": {
			S: aws.String(strconv.Itoa(meta.Size)),
		},
	}
}

func dynamoAssetAttrMapToMeta(m map[string]*dynamodb.AttributeValue) (meta AssetMeta) {
	d := map[string]string{
		"ObjID": "",  //ASSET_{ID}
		"AssetName": "",
		"Size": "0",
	}
	if err := dynamodbattribute.UnmarshalMap(m, &d); err != nil {
		log.WithFields(log.Fields{
			"context": "dynamoAssetAttrMapToMeta",
			"map": m,
		}).Error(err)
		return
	}
	meta.ID = strings.Replace(d["ObjID"], ASSET_KEY_PREFIX, "", 1)
	meta.Name = d["AssetName"]
	meta.Size, _ = strconv.Atoi(d["Size"])
	return meta
}

//metaTokenToDynamoAttrMap takes row with pk of TOKEN-{id} and returns the expiry and
//Asset Id associated
func dynamoTokenAttrMapToAssetToken(m map[string]*dynamodb.AttributeValue) (token AssetToken) {
	d := map[string]string{
		"AssetID": "", //Asset token refers to
		"ObjID": "",   //TOKEN_{id}
		"ObjSort": "0", //Token Expiry
	}
	if err := dynamodbattribute.UnmarshalMap(m, &d); err != nil {
		log.WithFields(log.Fields{
			"context": "dynamoTokenAttrMapToAssetMeta",
			"map": m,
		}).Error(err)
		return
	}
	token.AssetID = d["AssetID"]
	token.Token = strings.Replace(d["ObjID"], TOKEN_KEY_PREFIX, "", 1)
	if expiry, err := strconv.Atoi(d["ObjSort"], ); err == nil {
		token.Expiry = int64(expiry)
	}
	return token
}

func assetTokenToDynamoAttrMap(token AssetToken) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		"ObjID": {
			S:  aws.String(TOKEN_KEY_PREFIX + token.Token),
		},
		"ObjSort": {
			S: aws.String(strconv.Itoa(int(token.Expiry))),
		},
		"AssetID": {
			S: aws.String(token.AssetID),
		},
	}
}