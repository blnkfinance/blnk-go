package blnkgo

import (
	"fmt"
	"net/http"
	"time"
)

type IdentityService service

type Identity struct {
	IdentityID       string                 `json:"identity_id,omitempty"`
	IdentityType     IdentityType           `json:"identity_type"`
	FirstName        string                 `json:"first_name,omitempty"`
	LastName         string                 `json:"last_name,omitempty"`
	OtherNames       string                 `json:"other_names,omitempty"`
	Gender           string                 `json:"gender,omitempty"`
	DOB              *time.Time             `json:"dob,omitempty"`
	EmailAddress     string                 `json:"email_address"`
	PhoneNumber      string                 `json:"phone_number"`
	Nationality      string                 `json:"nationality,omitempty"`
	OrganizationName string                 `json:"organization_name,omitempty"`
	Category         string                 `json:"category"`
	Street           string                 `json:"street"`
	Country          string                 `json:"country"`
	State            string                 `json:"state"`
	PostCode         string                 `json:"post_code"`
	City             string                 `json:"city"`
	MetaData         map[string]interface{} `json:"meta_data,omitempty"`
}

type IdentityResponse struct {
	IdentityId string `json:"identity_id"`
	CreatedAt  string `json:"created_at"`
	Identity
}

// TokenizableIdentityField is a PascalCase Core struct field name (not snake_case JSON key).
type TokenizableIdentityField string

const (
	TokenizableFieldFirstName    TokenizableIdentityField = "FirstName"
	TokenizableFieldLastName     TokenizableIdentityField = "LastName"
	TokenizableFieldOtherNames   TokenizableIdentityField = "OtherNames"
	TokenizableFieldEmailAddress TokenizableIdentityField = "EmailAddress"
	TokenizableFieldPhoneNumber  TokenizableIdentityField = "PhoneNumber"
	TokenizableFieldStreet       TokenizableIdentityField = "Street"
	TokenizableFieldPostCode     TokenizableIdentityField = "PostCode"
)

// TokenizeFieldResponse is returned when a single identity field is tokenized.
type TokenizeFieldResponse struct {
	Message string `json:"message"`
}

func (s *IdentityService) Create(identity Identity) (*IdentityResponse, *http.Response, error) {
	//validate the identity
	if err := ValidateCreateIdentity(identity); err != nil {
		return nil, nil, err
	}
	identityResponse := new(IdentityResponse)
	req, err := s.client.NewRequest("identities", http.MethodPost, identity)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.CallWithRetry(req, identityResponse)
	if err != nil {
		return nil, resp, err
	}
	return identityResponse, resp, nil
}

func (s *IdentityService) Get(identityId string) (*IdentityResponse, *http.Response, error) {
	identityResponse := new(IdentityResponse)
	u := fmt.Sprintf("identities/%s", identityId)
	req, err := s.client.NewRequest(u, http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.CallWithRetry(req, identityResponse)
	if err != nil {
		return nil, resp, err
	}
	return identityResponse, resp, nil
}

func (s *IdentityService) List() ([]*IdentityResponse, *http.Response, error) {
	var identityResponse []*IdentityResponse
	req, err := s.client.NewRequest("identities", http.MethodGet, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.CallWithRetry(req, &identityResponse)
	if err != nil {
		return nil, resp, err
	}
	return identityResponse, resp, nil
}

func (s *IdentityService) Update(identityId string, identity *Identity) (*IdentityResponse, *http.Response, error) {
	var identityResponse *IdentityResponse
	u := fmt.Sprintf("identities/%s", identityId)
	req, err := s.client.NewRequest(u, http.MethodPut, identity)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.client.CallWithRetry(req, &identityResponse)
	if err != nil {
		return nil, resp, err
	}
	return identityResponse, resp, nil
}

func (s *IdentityService) Filter(params FilterParams) (*FilterResponse, *http.Response, error) {
	req, err := s.client.NewRequest("identities/filter", http.MethodPost, params)
	if err != nil {
		return nil, nil, err
	}

	var filterResponse FilterResponse
	resp, err := s.client.CallWithRetry(req, &filterResponse)
	if err != nil {
		return nil, resp, err
	}

	return &filterResponse, resp, nil
}

// TokenizeField tokenizes a single PII field on an identity.
func (s *IdentityService) TokenizeField(identityID string, field string) (*TokenizeFieldResponse, *http.Response, error) {
	if err := ValidateTokenizeIdentityField(identityID, field); err != nil {
		return nil, nil, err
	}

	u := fmt.Sprintf("identities/%s/tokenize/%s", identityID, field)
	req, err := s.client.NewRequest(u, http.MethodPost, nil)
	if err != nil {
		return nil, nil, err
	}

	tokenizeResp := new(TokenizeFieldResponse)
	resp, err := s.client.CallWithRetry(req, tokenizeResp)
	if err != nil {
		return nil, resp, err
	}

	return tokenizeResp, resp, nil
}

func NewIdentityService(client ClientInterface) *IdentityService {
	return &IdentityService{client: client}
}
