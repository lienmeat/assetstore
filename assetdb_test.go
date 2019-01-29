package assetstore

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupAssetDB() *DynamoDBMetaTokenStore {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return NewDynamoDBMetaTokenStore(os.Getenv("DYNAMODB_TABLE"), sess)
}

func TestDynamoDBMetaTokenStore_GetMeta(t *testing.T) {
	s := setupAssetDB()

	expect := AssetMeta{
		ID:      uuid.New().String(),
		Name:    "file.txt",
		Size:    500,
		Version: 0,
	}

	err := s.StoreMeta(expect)
	assert.NoError(t, err)

	type args struct {
		id string
	}
	tests := []struct {
		name     string
		args     args
		wantMeta AssetMeta
		wantErr  bool
	}{
		{
			name: "get meta",
			args: args{
				id: expect.ID,
			},
			wantMeta: expect,
			wantErr:  false,
		},
		{
			name: "non-existant meta",
			args: args{
				id: "dkfjajfukafjkajkfajkf",
			},
			wantMeta: AssetMeta{},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMeta, err := s.GetMeta(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DynamoDBMetaTokenStore.GetMeta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMeta, tt.wantMeta) {
				t.Errorf("DynamoDBMetaTokenStore.GetMeta() = %v, want %v", gotMeta, tt.wantMeta)
			}
		})
	}
}

func TestDynamoDBMetaTokenStore_StoreMeta(t *testing.T) {
	s := setupAssetDB()

	type args struct {
		meta AssetMeta
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantMeta AssetMeta
	}{
		{
			name: "store meta ok",
			args: args{
				meta: AssetMeta{
					ID:      "store meta id",
					Name:    "something.txt",
					Size:    400,
					Version: 0,
				},
			},
			wantErr: false,
			wantMeta: AssetMeta{
				ID:      "store meta id",
				Name:    "something.txt",
				Size:    400,
				Version: 0,
			},
		},
		{
			name: "no id",
			args: args{
				meta: AssetMeta{
					ID:      "",
					Name:    "something.txt",
					Size:    400,
					Version: 0,
				},
			},
			wantErr:  true,
			wantMeta: AssetMeta{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.StoreMeta(tt.args.meta)
			if (err != nil) != tt.wantErr {
				t.Errorf("DynamoDBMetaTokenStore.StoreMeta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(tt.args.meta.ID) != 0 {

				gotMeta, err := s.GetMeta(tt.args.meta.ID)
				assert.NoError(t, err)
				if !reflect.DeepEqual(gotMeta, tt.wantMeta) {
					t.Errorf("DynamoDBMetaTokenStore.StoreMeta() = %v, want %v", gotMeta, tt.wantMeta)
				}
			}
		})
	}
}

func TestDynamoDBMetaTokenStore_GetToken(t *testing.T) {
	s := setupAssetDB()

	expect := AssetToken{
		AssetID: uuid.New().String(),
		Token:   uuid.New().String(),
		Expiry:  time.Now().Add(time.Minute * 5).Unix(),
	}

	expired := AssetToken{
		AssetID: uuid.New().String(),
		Token:   uuid.New().String(),
		Expiry:  time.Now().Add(-time.Minute * 5).Unix(),
	}

	err := s.StoreToken(expect)
	assert.NoError(t, err)

	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		wantT   AssetToken
		wantErr bool
	}{
		{
			name: "get token ok",
			args: args{
				token: expect.Token,
			},
			wantT:   expect,
			wantErr: false,
		},
		{
			name: "expired token",
			args: args{
				token: expired.Token,
			},
			wantT:   AssetToken{},
			wantErr: true,
		},
		{
			name: "non existent token",
			args: args{
				token: "no token",
			},
			wantT:   AssetToken{},
			wantErr: true,
		},
		{
			name: "empty token",
			args: args{
				token: "",
			},
			wantT:   AssetToken{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotT, err := s.GetToken(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("DynamoDBMetaTokenStore.GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotT, tt.wantT) {
				t.Errorf("DynamoDBMetaTokenStore.GetToken() = %v, want %v", gotT, tt.wantT)
			}
		})
	}
}

func TestDynamoDBMetaTokenStore_StoreToken(t *testing.T) {
	s := setupAssetDB()

	okToken := AssetToken{
		Token:   uuid.New().String(),
		Expiry:  time.Now().Add(time.Minute * 5).Unix(),
		AssetID: uuid.New().String(),
	}

	type args struct {
		token AssetToken
	}

	tests := []struct {
		name      string
		args      args
		wantToken AssetToken
		wantErr   bool
	}{
		{
			name: "token store ok",
			args: args{
				token: okToken,
			},
			wantToken: okToken,
			wantErr:   false,
		},
		{
			name: "no asset id",
			args: args{
				token: AssetToken{
					Token:   "adfafaf",
					Expiry:  okToken.Expiry,
					AssetID: "",
				},
			},
			wantToken: AssetToken{},
			wantErr:   true,
		},
		{
			name: "no token",
			args: args{
				token: AssetToken{
					Token:   "",
					Expiry:  okToken.Expiry,
					AssetID: "kjdjafjaf",
				},
			},
			wantToken: AssetToken{},
			wantErr:   true,
		},
		{
			name: "expired",
			args: args{
				token: AssetToken{
					Token:   "jdkfjkafjaf",
					Expiry:  time.Now().Unix() - 5,
					AssetID: "kjdjafjaf",
				},
			},
			wantToken: AssetToken{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := s.StoreToken(tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("DynamoDBMetaTokenStore.StoreToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.args.token.Valid() {
				gotToken, err := s.GetToken(tt.args.token.Token)
				assert.NoError(t, err)
				if !reflect.DeepEqual(gotToken, tt.wantToken) {
					t.Errorf("DynamoDBMetaTokenStore.StoreToken() = %v, want %v", gotToken, tt.wantToken)
				}
			}
		})
	}
}