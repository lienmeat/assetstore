package assetstore

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//many adapters/backends might exist for storing asset contents/files
//s3 will do

type S3Storage struct {
	sess *session.Session
	bucket string
}

func NewS3Storage(bucket string, sess *session.Session) *S3Storage {
	return &S3Storage{
		sess: sess,
		bucket: bucket,
	}
}

func (s *S3Storage) Reader(id string) (reader io.ReadCloser, err error) {
	c := s3.New(s.sess)
	obj, err := c.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key: aws.String(id),
	})
	reader = obj.Body
	if err != nil || reader == nil {
		reader = ioutil.NopCloser(bytes.NewReader([]byte{}))
	}
	return reader, err
}

func (s *S3Storage) Writer(id string, reader io.ReadCloser) (n int64, err error) {
	defer reader.Close()
	uploader := s3manager.NewUploader(s.sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key: aws.String(id),
		Body: reader,
	})
	if err != nil {
		return 0, err
	}
	c := s3.New(s.sess)
	head, err := c.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key: aws.String(id),
	})
	return *head.ContentLength, err
}


