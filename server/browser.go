package server

import (
	"context"
	"log"
	"os"

	"github.com/ag-computational-bio/BioDataDBModels/go/client"
	"github.com/gin-gonic/gin"
)

//BrowserEndpoints Endpoints for the browser
type BrowserEndpoints struct {
	DataHandler DataHandler
	AutHandler  AuthHandler
	Token       string
}

//Browser structure for an igv browser
type Browser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Locus     string    `json:"locus"`
	Reference Reference `json:"reference"`
	Tracks    []Track   `json:"tracks"`
}

//ID a generic reusable ID
type ID struct {
	ID string `uri:"id" binding:"required"`
}

//GetDefaultTrackConfig
func (browser *BrowserEndpoints) GetDefaultTrackConfig(c *gin.Context) {
	token := browser.AutHandler.GetAccessTokenFromGinContext(c)

	token = os.Getenv("APIToken")

	currentRefFastaVersion, err := browser.DataHandler.getCurrentDatasetVersion(FastaRef, token)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithError(400, err)
		return
	}

	refFiles, err := browser.DataHandler.getDatasetDownloadLinks(currentRefFastaVersion, token)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithError(400, err)
		return
	}

	currentAnnotationGffVersion, err := browser.DataHandler.getCurrentDatasetVersion(GffRef, token)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithError(400, err)
		return
	}

	gffAnnotationFiles, err := browser.DataHandler.getDatasetDownloadLinks(currentAnnotationGffVersion, token)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithError(400, err)
		return
	}

	gffTrack := Track{
		Type:       "annotation",
		Format:     "gff3",
		Name:       "Annotation",
		AutoHeight: true,
		URL:        gffAnnotationFiles.GetLinks()[0].GetLink()[0],
	}

	reference := Reference{
		Name:     "NC_002942",
		ID:       "NC_002942",
		FastaURL: refFiles.GetLinks()[0].GetLink()[0],
		IndexURL: refFiles.GetLinks()[0].GetLink()[1],
		Tracks:   []Track{gffTrack},
	}

	igv_browser := Browser{
		ID:        "NC_002942",
		Name:      "NC_002942",
		Reference: reference,
		Tracks:    make([]Track, 0),
	}

	c.JSON(200, igv_browser)

}

func (browser *BrowserEndpoints) GetBigWigsTracks(c *gin.Context) {
	var id ID
	err := c.BindUri(&id)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithError(400, err)
		return
	}

	token := browser.AutHandler.GetAccessTokenFromGinContext(c)

	token = os.Getenv("APIToken")

	tracks, err := browser.DataHandler.GetBigWigsTrack(id.ID, token)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithError(400, err)
		return
	}

	c.JSON(200, tracks)
}

func (browser *BrowserEndpoints) GetBamTrack(c *gin.Context) {
	var id ID
	err := c.BindUri(&id)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithError(400, err)
		return
	}

	token := browser.AutHandler.GetAccessTokenFromGinContext(c)

	token = os.Getenv("APIToken")

	tracks, err := browser.DataHandler.GetBamTrack(id.ID, token)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithError(400, err)
		return
	}

	c.JSON(200, tracks)
}

//IGVBrowser Starts the igv viewer
func (browser *BrowserEndpoints) IGVBrowser(c *gin.Context) {
	token := browser.AutHandler.GetAccessTokenFromGinContext(c)

	token = os.Getenv("APIToken")

	bigWigsList, err := browser.DataHandler.GetBigWigsList(token)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithError(400, err)
		return
	}

	bamList, err := browser.DataHandler.GetBamList(token)
	if err != nil {
		log.Println(err.Error())
		c.AbortWithError(400, err)
		return
	}

	c.HTML(200, "browser.html", gin.H{"BigWigsList": bigWigsList, "BamList": bamList})
}

func (browser *BrowserEndpoints) outgoingContext() context.Context {
	return browser.DataHandler.GRPCEndpoints.OutGoingContextFromToken(browser.Token, client.UserAPIToken)
}
