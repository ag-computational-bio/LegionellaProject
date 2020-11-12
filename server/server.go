package server

import (
	"fmt"
	"log"
	"strconv"

	"github.com/ag-computational-bio/BioDataDBModels/go/client"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"

	"github.com/spf13/viper"
)

//Run Starts the webserver and reads the config
func Run() {
	//Load the required config
	host := viper.GetString("Endpoints.DatasetHandler.Host")
	if host == "" {
		err := fmt.Errorf("Endpoints datasethandler host needs to be set")
		log.Fatalln(err.Error())
	}

	port := viper.GetInt("Endpoints.DatasetHandler.Port")
	if port == 0 {
		err := fmt.Errorf("Endpoints datasethandler port needs to be set")
		log.Fatalln(err.Error())
	}

	//The ids are from the BioDataDB
	bigWigsID := viper.GetString("Datasets.Bigwigs")
	bamID := viper.GetString("Datasets.Bam")
	referenceID := viper.GetString("Datasets.Reference")
	gffID := viper.GetString("Datasets.GFFAnnotation")

	//Establish the grpc client to connect to the BioDataDB
	grpcClients := client.GRPCEndpointsClients{}
	err := grpcClients.New(host, strconv.Itoa(port))
	if err != nil {
		log.Fatalln(err.Error())
	}

	//Init the authhandler
	//Will only be used until the publication of the project
	authhandler := AuthHandler{}
	authhandler.Init()

	datahandler := DataHandler{
		AutHandler:        authhandler,
		BWDatsetID:        bigWigsID,
		BamDatasetID:      bamID,
		FastaRefDatasetID: referenceID,
		GFFDatasetID:      gffID,
		GRPCEndpoints:     grpcClients,
	}

	browserEndpoints := BrowserEndpoints{
		DataHandler: datahandler,
		AutHandler:  authhandler,
	}

	router := gin.Default()

	router.HTMLRender = createMyRender()
	router.Static("./static", "./static")

	//Temporary solution until final publication to restric access
	router.Use(authhandler.UpdateToken)

	router.GET("/", base)
	router.GET("/index", index)
	router.GET("/auth/callback", authhandler.Callback)
	router.GET("/login", authhandler.Auth)

	dataGroup := router.Group("/data")
	dataGroup.GET("/default", browserEndpoints.GetDefaultTrackConfig)
	dataGroup.GET("/bigWigsTrack/:id", browserEndpoints.GetBigWigsTracks)

	browserGroup := router.Group("/browser")
	browserGroup.GET("/", browserEndpoints.IGVBrowser)

	router.Run()
}

func createMyRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	r.AddFromFiles("index.html", "templates/index.html", "templates/baseTopBar.html", "templates/baseHeader.html")
	r.AddFromFiles("browser.html", "templates/browser.html", "templates/baseTopBar.html", "templates/baseHeader.html")

	return r
}
