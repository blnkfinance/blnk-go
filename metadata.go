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

// Metadata is the response of POST /{entity-id}/metadata. Core returns the
// merged metadata under "meta_data".
type Metadata struct {
	MetaData MetaData `json:"meta_data"`
}

func (s *MetadataService) UpdateMetadata(entityID string, body UpdateMetaDataRequest) (*Metadata, *http.Response, error) {
	if entityID == "" {
		return nil, nil, fmt.Errorf("entity ID is required")
	}

	u, err := endpointPath("%s/metadata", entityID)
	if err != nil {
		return nil, nil, err
	}

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
