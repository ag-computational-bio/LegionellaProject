package server

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/ag-computational-bio/BioDataDBModels/go/datasetapimodels"

	"github.com/ag-computational-bio/BioDataDBModels/go/client"
	"github.com/ag-computational-bio/BioDataDBModels/go/commonmodels"
	"github.com/ag-computational-bio/BioDataDBModels/go/datasetentrymodels"
	"github.com/ag-computational-bio/BioDataDBModels/go/loadmodels"
)

//TrackType Supported IGV track file format, associated track types can be found here: https://github.com/igvteam/igv.js/wiki/Tracks-2.0
type TrackType string

const (
	BigWigs  TrackType = "BigWigs"
	BAM      TrackType = "BAM"
	FastaRef TrackType = "FASTA"
	GffRef   TrackType = "GFF"
)

// Reference structure for the genome reference
type Reference struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	FastaURL string  `json:"fastaURL"`
	IndexURL string  `json:"indexURL"`
	Tracks   []Track `json:"tracks"`
}

//Track structure for an igv track, not all fields are always required
type Track struct {
	Name       string      `json:"name,omitempty"`
	URL        string      `json:"url,omitempty"`
	Format     string      `json:"format,omitempty"`
	IndexURL   string      `json:"indexURL,omitempty"`
	Type       string      `json:"type,omitempty"`
	Min        int         `json:"min,omitempty"`
	Max        int         `json:"max,omitempty"`
	AutoScale  bool        `json:"autoscale,omitempty"`
	Color      string      `json:"color,omitempty"`
	Indexed    string      `json:"indexed,omitempty"`
	AutoHeight bool        `json:"autoHeight,omitempty"`
	Searchable bool        `json:"searchable,omitempty"`
	GuideLines []GuideLine `json:"guidelines,omitempty"`
}

//GuideLine https://github.com/igvteam/igv.js/wiki/Wig-Track
type GuideLine struct {
	Color  string `json:"color,omitempty"`
	Dotted bool   `json:"dotted,omitempty"`
	Y      int    `json:"y,omitempty"`
}

//DataHandler Handles the data connection with the BioDataDB backend
type DataHandler struct {
	GRPCEndpoints     client.GRPCEndpointsClients
	AutHandler        AuthHandler
	BWDatsetID        string
	BamDatasetID      string
	FastaRefDatasetID string
	GFFDatasetID      string
}

//FileData Stores a structed set of filesgroups, can be used to subdivide the dropdown menu
type FileData struct {
	BAMData     map[string][]FileGroup
	BigWigsData map[string][]FileGroup
}

//FileGroup A group of files (reference a BioDataDB ObjectGroup)
type FileGroup struct {
	GroupID   string
	GroupName string
	Objects   []FileDescription
}

//FileDescription A reference to an BioDataDB object
type FileDescription struct {
	Name string
	ID   string
}

func (datahandler *DataHandler) GetBamList(token string) (map[string][]FileGroup, error) {
	datasetVersion, err := datahandler.getCurrentDatasetVersion(BAM, token)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	groupList, err := datahandler.getDatasetObjectGroupList(BAM, datasetVersion, token)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	bamList := make(map[string][]FileGroup)
	var fileGroupData []FileGroup

	for _, objectGroup := range groupList.DatasetObjectGroups {

		if len(objectGroup.GetObjects()) < 1 {
			log.Println(fmt.Sprintf("ObjectGroup with id: %v and name: %v has no associated objects", objectGroup.GetID(), objectGroup.GetName()))
			continue
		}

		objectGroupRepr := FileGroup{
			GroupID:   objectGroup.GetID(),
			GroupName: objectGroup.GetObjects()[0].GetFilename(),
		}

		fileGroupData = append(fileGroupData, objectGroupRepr)

	}

	bamList["ALL"] = fileGroupData

	return bamList, nil
}

//GetBamTrack Returns a bam track with a specific id with the default config
func (datahandler *DataHandler) GetBamTrack(id string, token string) ([]Track, error) {
	objectGroup, err := datahandler.getObjectGroup(id, token)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	var tracks []Track

	for _, objectGroup := range objectGroup.GetLinks() {
		objects := objectGroup.GetObject().GetObjects()
		if len(objects) < 2 {
			err := errors.New("Not enough objects in object group for bam file")
			log.Println(err)
		}

		track := Track{
			Color:     "rgb(0, 0, 150)",
			AutoScale: true,
			Type:      "alignment",
			Format:    "bam",
		}

		for i, object := range objects {
			if strings.HasSuffix(object.Filename, ".bam") {
				track.Name = object.GetFilename()
				track.URL = objectGroup.Link[i]
			} else if strings.HasSuffix(object.Filename, ".bam.bai") {
				track.IndexURL = objectGroup.Link[i]
			}
		}
		tracks = append(tracks, track)
	}

	return tracks, nil

}

//GetBigWigsTrack Returns a bigwigs track with a specific id with the default config
func (datahandler *DataHandler) GetBigWigsTrack(id string, token string) ([]Track, error) {
	objectGroup, err := datahandler.getObjectGroup(id, token)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	var tracks []Track

	for _, objects := range objectGroup.GetLinks() {
		for i, object := range objects.GetObject().GetObjects() {
			track := Track{
				Color:     "rgb(0, 0, 150)",
				AutoScale: true,
				Type:      "wig",
				Name:      object.GetFilename(),
				URL:       objects.GetLink()[i],
			}

			tracks = append(tracks, track)
		}
	}

	return tracks, nil

}

//GetBigWigsList List of bigwigs file grouped by forward and reverse files
func (datahandler *DataHandler) GetBigWigsList(token string) (map[string][]FileGroup, error) {
	datasetVersion, err := datahandler.getCurrentDatasetVersion(BigWigs, token)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	groupList, err := datahandler.getDatasetObjectGroupList(BigWigs, datasetVersion, token)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	bigWigsList := make(map[string][]FileGroup)
	var fileGroupData []FileGroup

	for _, objectGroup := range groupList.DatasetObjectGroups {
		objectGroupRepr := FileGroup{
			GroupID:   objectGroup.GetID(),
			GroupName: objectGroup.GetName(),
			Objects:   make([]FileDescription, 0),
		}
		for _, object := range objectGroup.GetObjects() {
			splitedString := strings.Split(object.GetFilename(), "_")
			correctedSplittedString := splitedString[:len(splitedString)-1]
			groupName := strings.Join(correctedSplittedString, "_")

			objectGroupRepr.GroupName = groupName
			object := FileDescription{
				ID:   object.GetID(),
				Name: object.GetFilename(),
			}

			objectGroupRepr.Objects = append(objectGroupRepr.Objects, object)
		}

		fileGroupData = append(fileGroupData, objectGroupRepr)
	}

	bigWigsList["ALL"] = fileGroupData

	return bigWigsList, nil
}

//getCurrentDatasetVersion Returns the current DatasetVersion of the dataset for a specific type of track
func (datahandler *DataHandler) getCurrentDatasetVersion(trackType TrackType, token string) (*datasetentrymodels.DatasetVersionEntry, error) {
	var id string
	switch trackType {
	case BigWigs:
		id = datahandler.BWDatsetID
	case BAM:
		id = datahandler.BamDatasetID
	case FastaRef:
		id = datahandler.FastaRefDatasetID
	case GffRef:
		id = datahandler.GFFDatasetID
	}

	datasetID := commonmodels.ID{
		ID: id,
	}

	datasetVersion, err := datahandler.GRPCEndpoints.DatasetBackend.GetCurrentVersionOfDataset(datahandler.AutHandler.OutGoingContextFromToken(token, client.UserAPIToken), &datasetID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return datasetVersion, nil
}

//getDatasetObjectGroupList Returns all object groups of a specific dataset version
func (datahandler *DataHandler) getDatasetObjectGroupList(trackType TrackType, datasetVersion *datasetentrymodels.DatasetVersionEntry, token string) (*datasetapimodels.DatasetObjectGroupList, error) {
	datasetVersionID := commonmodels.ID{
		ID: datasetVersion.GetID(),
	}

	datasetObjects, err := datahandler.GRPCEndpoints.DatasetBackend.DatasetVersionObjectGroups(datahandler.AutHandler.OutGoingContextFromToken(token, client.UserAPIToken), &datasetVersionID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return datasetObjects, nil
}

//getDatasetDownloadLinks Returns presigned download urls for a specific dataset version
func (datahandler *DataHandler) getDatasetDownloadLinks(datasetVersion *datasetentrymodels.DatasetVersionEntry, token string) (*loadmodels.GetDownloadResponse, error) {
	var requests []*loadmodels.ResourceRequest

	requests = append(requests, &loadmodels.ResourceRequest{
		Resource:   commonmodels.Resource_DatasetVersion,
		ResourceID: datasetVersion.GetID(),
	})

	downloadRequest := loadmodels.GetDownloadRequest{
		Resource: requests,
	}

	groupLinks, err := datahandler.GRPCEndpoints.LoadBackend.GetDownloadLinks(datahandler.AutHandler.OutGoingContextFromToken(token, client.UserAPIToken), &downloadRequest)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return groupLinks, nil
}

//getObjectGroup Returns all presigned download links for a specific dataset version
func (datahandler *DataHandler) getObjectGroup(groupID string, token string) (*loadmodels.GetDownloadResponse, error) {
	var requests []*loadmodels.ResourceRequest

	requests = append(requests, &loadmodels.ResourceRequest{
		Resource:   commonmodels.Resource_DatasetObjectGroupResource,
		ResourceID: groupID,
	})

	downloadRequest := loadmodels.GetDownloadRequest{
		Resource: requests,
	}

	groupLinks, err := datahandler.GRPCEndpoints.LoadBackend.GetDownloadLinks(datahandler.AutHandler.OutGoingContextFromToken(token, client.UserAPIToken), &downloadRequest)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return groupLinks, nil
}
