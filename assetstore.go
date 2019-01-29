package assetstore

import (
	"io"
	"time"

	log "github.com/sirupsen/logrus"
)

//High level types and interfaces for asset storage

//Properties our assets might have
type AssetMeta struct {
	ID string `json:"id"`
	//File or asset name
	Name string `json:"name"`
	//Size in bytes
	Size int    `json:"size"`
	//Version of asset
	Version int `json:"version"`
}

func (m AssetMeta) Valid() bool {
	return m.ID != "" && m.Name != ""
}

type AssetToken struct {
	Token string `json:"token,omitempty"`
	//Expiry unix timestamp
	Expiry int64 `json:"expiry,omitempty"`
	//AssetID
	AssetID string `json:"asset_id,omitempty"`
}

func (t AssetToken) Valid() bool {
	return t.AssetID != "" && t.Token != "" && time.Now().Before(time.Unix(t.Expiry, 0))
}

//AssetIDRetriever retrieves an assets' io.ReadCloser and meta by its id
type AssetIDRetriever interface {
	GetByID(id string) (meta AssetMeta, asset io.ReadCloser, err error)
}

//AssetTokenRetriever retrieves an assets' io.ReadCloser and meta by via a token
type AssetTokenRetriever interface {
	GetByToken(token string) (meta AssetMeta, asset io.ReadCloser, err error)
}

//Stores an asset given its meta, token, and a io.ReadCloser
type AssetStorer interface {
	Store(meta AssetMeta, token AssetToken, asset io.ReadCloser) (err error)
}

type MetaRetriever interface {
	GetMeta(id string) (meta AssetMeta, err error)
}

type MetaStorer interface {
	StoreMeta(meta AssetMeta) (err error)
}

type TokenRetriever interface {
	GetToken(token string) (t AssetToken, err error)
}

type TokenStorer interface {
	StoreToken(token AssetToken) (err error)
}

type AssetDataReader interface {
	Reader(id string) (reader io.ReadCloser, err error)
}

type AssetDataWriter interface {
	Writer(id string, reader io.ReadCloser) (n int64, err error)
}

type AssetDataHandler interface{
	AssetDataReader
	AssetDataWriter
}

type AssetMetaHandler interface {
	MetaRetriever
	MetaStorer
}

type AssetTokenHandler interface {
	TokenRetriever
	TokenStorer
}

type AssetHandler interface {
	AssetIDRetriever
	AssetTokenRetriever
	AssetStorer
}

type AssetStorage struct {
	metaHandler AssetMetaHandler
	tokenHandler AssetTokenHandler
	dataHandler AssetDataHandler
}

func NewAssetStorage(metaHandler AssetMetaHandler, tokenHandler AssetTokenHandler, dataHandler AssetDataHandler) *AssetStorage {
	return &AssetStorage{
		metaHandler: metaHandler,
		tokenHandler: tokenHandler,
		dataHandler: dataHandler,
	}
}

func (s *AssetStorage) GetByID(id string) (meta AssetMeta, asset io.ReadCloser, err error) {
	meta, err = s.metaHandler.GetMeta(id)
	if err != nil {
		log.WithFields(log.Fields{
			"context": "AssetStorage.GetByID()",
			"id": id,
			"metaHandler": s.metaHandler,
			"meta": meta,
		}).Error(err)
		return
	}
	asset, err = s.dataHandler.Reader(id)
	if err != nil {
		log.WithFields(log.Fields{
			"context": "AssetStorage.GetByID()",
			"id": id,
			"metaHandler": s.metaHandler,
			"meta": meta,
		}).Error(err)
		return
	}
	return
}

func (s *AssetStorage) GetByToken(token string) (meta AssetMeta, asset io.ReadCloser, err error) {
	aToken, err := s.tokenHandler.GetToken(token)
	if err != nil {
		log.WithFields(log.Fields{
			"context": "AssetStorage.GetByToken()",
			"tokenHandler": s.tokenHandler,
			"token": token,
		}).Error(err)
		return
	}
	meta, err = s.metaHandler.GetMeta(aToken.AssetID)
	if err != nil {
		log.WithFields(log.Fields{
			"context": "AssetStorage.GetByToken()",
			"tokenHandler": s.tokenHandler,
			"token": token,
			"meta": meta,
		}).Error(err)
		return
	}
	asset, err = s.dataHandler.Reader(meta.ID)
	if err != nil {
		log.WithFields(log.Fields{
			"context": "AssetStorage.GetByToken()",
			"tokenHandler": s.tokenHandler,
			"dataHandler": s.dataHandler,
			"token": token,
			"meta": meta,
		}).Error(err)
		return
	}
	return
}

func (s *AssetStorage) Store(meta AssetMeta, token AssetToken, asset io.ReadCloser) (err error) {
	n, err := s.dataHandler.Writer(meta.ID, asset)
	if err != nil {
		log.WithFields(log.Fields{
			"context": "AssetStorage.Store()",
			"dataHandler": s.dataHandler,
			"token": token,
			"meta": meta,
		}).Error(err)
		return
	}
	meta.Size = int(n)
	err = s.metaHandler.StoreMeta(meta)
	if err != nil {
		log.WithFields(log.Fields{
			"context": "AssetStorage.Store()",
			"dataHandler": s.dataHandler,
			"metaHandler": s.metaHandler,
			"token": token,
			"meta": meta,
		}).Error(err)
		return
	}
	if token.Valid() {
		err = s.tokenHandler.StoreToken(token)
		if err != nil {
			log.WithFields(log.Fields{
				"context":     "AssetStorage.Store()",
				"dataHandler": s.dataHandler,
				"metaHandler": s.metaHandler,
				"tokenHandler": s.tokenHandler,
				"token":       token,
				"meta":        meta,
			}).Error(err)
		}
	}
	return
}