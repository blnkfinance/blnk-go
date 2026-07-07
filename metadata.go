package blnkgo

import (
	"fmt"
	"net/http"
)

// MetaData is a flexible key-value metadata map attached to Blnk resources.
type MetaData = map[string]interface{}

type MetadataService service

type UpdateMetaDataRequest struct {
	MetaData MetaData `json:"meta_data"`
}

type Metadata struct {
	MetaData MetaData `json:"metadata"`
}

func (s *MetadataService) UpdateMetadata(entityID string, body UpdateMetaDataRequest) (*Metadata, *http.Response, error) {
	if entityID == "" {
		return nil, nil, fmt.Errorf("entity ID is required")
	}

	u := fmt.Sprintf("%s/metadata", entityID)

	req, err := s.client.NewRequest(u, http.MethodPost, body)
	if err != nil {
		return nil, nil, err
	}

	response := new(Metadata)
	resp, err := s.client.CallWithRetry(req, response)
	if err != nil {
		return nil, resp, err
	}

	return response, resp, nil
}

func NewMetadataService(client ClientInterface) *MetadataService {
	return &MetadataService{client: client}
}
