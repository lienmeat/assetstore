package assetstore

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//idRetriever gets assets by id
var idRetriever AssetIDRetriever
//tokenRetriever gets assets by token
var tokenRetriever AssetTokenRetriever
//assetStorer stores assets given metadata and data
var assetStorer AssetStorer

// For uptime watchers
func ping(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func addAsset(c *gin.Context) {
	type input struct {
		Name string `json:"name" form:"name"`
		Token bool `json:"token" form:"token"`
		Expiry int `json:"expiry" form:"expiry"`
	}

	type addResp struct {
		Meta AssetMeta `json:"asset"`
		Token AssetToken `json:"token,omitempty"`
		Error string `json:"error"`
	}

	i := input{}

	c.Bind(&i)

	isForm := strings.Contains(strings.ToLower(c.ContentType()), "multipart")
	name := c.Param("assetname")
	if !isForm && name == "" {
		c.JSON(http.StatusBadRequest, addResp{Error: "asset name not specified"})
		return
	}else{
		i.Name = name
	}


	if i.Token == false {
		i.Expiry = 0
	}

	meta := AssetMeta{
		ID: uuid.New().String(),
		Name: i.Name,
		Size: int(c.Request.ContentLength),
		Version: 0,
	}

	token := AssetToken{}
	if i.Token && i.Expiry != 0 {
		//populate a new token
		token.Token = uuid.New().String()
		token.Expiry = time.Now().Add(time.Minute * time.Duration(i.Expiry)).Unix()
		token.AssetID = meta.ID
	}

	var reader io.ReadCloser

	if !isForm {
		reader = c.Request.Body
	}else {
		ff, err := c.FormFile("file")
		if err == nil {
			reader, _ = ff.Open()
			meta.Name = ff.Filename
			meta.Size = int(ff.Size)
		} else {
			c.JSON(http.StatusBadRequest, addResp{Error: err.Error()})
			return
		}
	}

	err := assetStorer.Store(meta, token, reader)
	if err != nil {
		c.JSON(http.StatusBadRequest, addResp{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, addResp{Meta: meta, Token: token})
}

func getAssetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, "asset id not specified")
		return
	}
	meta, asset, err := idRetriever.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNoContent, err.Error())
		return
	}
	defer asset.Close()
	sendAsset(c, asset, meta)
	return
}

func getAssetByToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, "token not specified")
		return
	}
	meta, asset, err := tokenRetriever.GetByToken(token)
	if err != nil {
		c.JSON(http.StatusNoContent, err.Error())
		return
	}
	defer asset.Close()
	sendAsset(c, asset, meta)
	return
}

//sendAsset transfers asset/file to the client as a download
func sendAsset(c *gin.Context, asset io.ReadCloser, meta AssetMeta) {
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+ meta.Name)
	c.Header("Content-Length", fmt.Sprintf("%d", meta.Size))
	c.DataFromReader(http.StatusOK, int64(meta.Size), "application/octet-stream", asset, map[string]string{})
}


func initCORS(server gin.IRouter) {
	corsconfig := cors.DefaultConfig()
	corsconfig.AllowAllOrigins = true
	corsconfig.AddAllowMethods([]string{"GET", "POST", "HEAD"}...)
	//supporting auth would be a next step
	corsconfig.AddAllowHeaders("authorization")
	corsconfig.AddAllowHeaders("x-api-key")
	server.Use(cors.New(corsconfig))
}

func RunAPI(idr AssetIDRetriever, tor AssetTokenRetriever, storer AssetStorer, basePath string, port int) {
	idRetriever = idr
	tokenRetriever = tor
	assetStorer = storer

	server := gin.Default()
	initCORS(server)
	base := server.Group(basePath)
	base.GET("/ping", ping)

	base.GET("/asset/:id", getAssetByID)
	base.GET("/asset-token/:token", getAssetByToken)
	base.POST("/asset", addAsset)
	base.POST("/asset/:assetname", addAsset)


	server.Run(fmt.Sprintf("0.0.0.0:%d", port))
}
