package assetstore

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupS3Storage() *S3Storage {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return NewS3Storage(os.Getenv("S3_BUCKET"), sess)
}

func TestS3Storage_Reader(t *testing.T) {

	s := setupS3Storage()
	fn := uuid.New().String()
	_, err := s.Writer(fn, ioutil.NopCloser(bytes.NewReader([]byte("here we go"))))
	assert.NoError(t, err)

	type args struct {
		id string
	}
	tests := []struct {
		name      string
		args      args
		wantBytes []byte
		wantErr   bool
	}{
		{
			name: "testfile",
			args: args{
				id: fn,
			},
			wantBytes: []byte("here we go"),
			wantErr:   false,
		},
		{
			name: "no such id",
			args: args{
				id: "no such id",
			},
			wantBytes: []byte(""),
			wantErr:   true,
		},
		{
			name: "no id",
			args: args{
				id: "",
			},
			wantBytes: []byte(""),
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReader, err := s.Reader(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("S3Storage.Reader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotBytes, err := ioutil.ReadAll(gotReader)
			assert.NoError(t, err)
			if !reflect.DeepEqual(gotBytes, tt.wantBytes) {
				t.Errorf("S3Storage.Reader() = %v, want %v", gotBytes, tt.wantBytes)
			}
		})
	}
}

func TestS3Storage_Writer(t *testing.T) {

	s := setupS3Storage()
	data := []byte("just some data")

	type args struct {
		id     string
		data   []byte
	}
	tests := []struct {
		name    string
		args    args
		wantN   int64
		wantErr bool
		wantData []byte
	}{
		{
			name: "some data",
			args: args{
				id: uuid.New().String(),
				data: data,
			},
			wantN: int64(len(data)),
			wantErr: false,
			wantData: data,
		},
		{
			name: "empty file",
			args: args{
				id: uuid.New().String(),
				data: []byte(""),
			},
			wantN: 0,
			wantErr: false,
			wantData: []byte(""),
		},
		{
			name: "no id",
			args: args{
				id: "",
				data: []byte("some data here"),
			},
			wantN: 0,
			wantErr: true,
			wantData: []byte(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := ioutil.NopCloser(bytes.NewReader(tt.args.data))
			gotN, err := s.Writer(tt.args.id, reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("S3Storage.Writer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("S3Storage.Writer() = %v, want %v", gotN, tt.wantN)
			}

			if len(tt.args.id) != 0 {
				reader, err = s.Reader(tt.args.id)
				assert.NoError(t, err)
				actual, err := ioutil.ReadAll(reader)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantData, actual)
			}
		})
	}
}
