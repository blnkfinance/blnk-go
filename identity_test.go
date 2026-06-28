package blnkgo_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupIdentityService() (*MockClient, *blnkgo.IdentityService) {
	mockClient := &MockClient{}
	svc := blnkgo.NewIdentityService(mockClient)
	return mockClient, svc
}

func TestIdentityService_Create(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identity := blnkgo.Identity{
		IdentityType: blnkgo.Individual,
		FirstName:    "John",
		LastName:     "Doe",
		EmailAddress: "john.doe@example.com",
		PhoneNumber:  "1234567890",
		Category:     "customer",
		Street:       "123 Main St",
		Country:      "USA",
		State:        "CA",
		PostCode:     "90001",
		City:         "Los Angeles",
		DOB:          &time.Time{},
		Gender:       "Male",
		Nationality:  "Nigerian",
	}

	t.Run("successful creation", func(t *testing.T) {
		expectedResponse := &blnkgo.IdentityResponse{
			IdentityId: "12345",
			CreatedAt:  time.Now().Format(time.RFC3339),
			Identity:   identity,
		}

		mockClient.On("NewRequest", "identities", http.MethodPost, identity).Return(&http.Request{}, nil)
		mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			resp := args.Get(1).(*blnkgo.IdentityResponse)
			*resp = *expectedResponse
		}).Return(&http.Response{}, nil)

		resp, httpResp, err := svc.Create(identity)
		assert.NoError(t, err)
		assert.NotNil(t, httpResp)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		invalidIdentity := blnkgo.Identity{
			IdentityID:   "not-a-valid-id",
			EmailAddress: "john.doe@example.com",
			PhoneNumber:  "1234567890",
		}

		resp, httpResp, err := svc.Create(invalidIdentity)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Nil(t, httpResp)
		assert.Contains(t, err.Error(), "idt_")
	})
}

func TestIdentityService_ServerError(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identity := blnkgo.Identity{
		IdentityType:     blnkgo.Organization,
		OrganizationName: "ACME Inc",
	}

	mockClient.On("NewRequest", "identities", http.MethodPost, identity).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, errors.New("server error"))

	resp, httpResp, err := svc.Create(identity)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusInternalServerError, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}
func TestIdentityService_Get(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityId := "12345"
	expectedResponse := &blnkgo.IdentityResponse{
		IdentityId: identityId,
		CreatedAt:  time.Now().Format(time.RFC3339),
		Identity: blnkgo.Identity{
			IdentityType: blnkgo.Individual,
			FirstName:    "John",
			LastName:     "Doe",
			EmailAddress: "john.doe@example.com",
			PhoneNumber:  "1234567890",
			Category:     "customer",
			Street:       "123 Main St",
			Country:      "USA",
			State:        "CA",
			PostCode:     "90001",
			City:         "Los Angeles",
			DOB:          &time.Time{},
			Gender:       "Male",
			Nationality:  "Nigerian",
		},
	}

	t.Run("successful get", func(t *testing.T) {
		mockClient.On("NewRequest", fmt.Sprintf("identities/%s", identityId), http.MethodGet, nil).Return(&http.Request{}, nil)
		mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			resp := args.Get(1).(*blnkgo.IdentityResponse)
			*resp = *expectedResponse
		}).Return(&http.Response{}, nil)

		resp, httpResp, err := svc.Get(identityId)
		assert.NoError(t, err)
		assert.NotNil(t, httpResp)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

func TestIdentityService_Get_NotFound(t *testing.T) {
	mockClient, svc := setupIdentityService()
	identityId := "12345"
	mockClient.On("NewRequest", fmt.Sprintf("identities/%s", identityId), http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusNotFound}, errors.New("not found"))

	resp, httpResp, err := svc.Get(identityId)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "not found")
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Get_ServerError(t *testing.T) {
	mockClient, svc := setupIdentityService()
	identityId := "12345"
	mockClient.On("NewRequest", fmt.Sprintf("identities/%s", identityId), http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, errors.New("server error"))

	resp, httpResp, err := svc.Get(identityId)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusInternalServerError, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "server error")
}

func TestIdentityService_List(t *testing.T) {
	mockClient, svc := setupIdentityService()

	t.Run("successful list", func(t *testing.T) {
		expectedResponse := []*blnkgo.IdentityResponse{
			{
				IdentityId: "12345",
				CreatedAt:  time.Now().Format(time.RFC3339),
				Identity: blnkgo.Identity{
					IdentityType: blnkgo.Individual,
					FirstName:    "John",
					LastName:     "Doe",
					EmailAddress: "john@example.com",
					PhoneNumber:  "1234567890",
					Category:     "customer",
					Street:       "123 Main St",
					Country:      "USA",
					State:        "CA",
					PostCode:     "90001",
					City:         "Los Angeles",
				},
			},
			{
				IdentityId: "67890",
				CreatedAt:  time.Now().Format(time.RFC3339),
				Identity: blnkgo.Identity{
					IdentityType:     blnkgo.Organization,
					OrganizationName: "ACME Inc",
					EmailAddress:     "contact@acme.com",
					PhoneNumber:      "0987654321",
					Category:         "business",
					Street:           "456 Corp Ave",
					Country:          "USA",
					State:            "NY",
					PostCode:         "10001",
					City:             "New York",
				},
			},
		}

		mockClient.On("NewRequest", "identities", http.MethodGet, nil).Return(&http.Request{}, nil)
		mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			resp := args.Get(1).(*[]*blnkgo.IdentityResponse)
			*resp = expectedResponse
		}).Return(&http.Response{}, nil)

		resp, httpResp, err := svc.List()
		assert.NoError(t, err)
		assert.NotNil(t, httpResp)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

func TestIdentityService_List_ServerError(t *testing.T) {
	mockClient, svc := setupIdentityService()

	mockClient.On("NewRequest", "identities", http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, errors.New("server error"))

	resp, httpResp, err := svc.List()
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusInternalServerError, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Update_Successful(t *testing.T) {
	mockClient, svc := setupIdentityService()
	identityId := "12345"

	updateIdentity := &blnkgo.Identity{
		FirstName:    "Jane",
		LastName:     "Doe",
		EmailAddress: "jane.doe@example.com",
		PhoneNumber:  "0987654321",
		Category:     "customer",
		Street:       "456 Oak St",
		Country:      "USA",
		State:        "NY",
		PostCode:     "10001",
		City:         "New York",
	}

	expectedResponse := &blnkgo.IdentityResponse{
		IdentityId: identityId,
		CreatedAt:  time.Now().Format(time.RFC3339),
		Identity:   *updateIdentity,
	}

	mockClient.On("NewRequest", fmt.Sprintf("identities/%s", identityId), http.MethodPut, updateIdentity).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		resp := args.Get(1).(**blnkgo.IdentityResponse)
		*resp = expectedResponse
	}).Return(&http.Response{}, nil)

	resp, httpResp, err := svc.Update(identityId, updateIdentity)
	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, expectedResponse, resp)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Update_ClearsFieldsWithEmptyStrings(t *testing.T) {
	mockClient, svc := setupIdentityService()
	identityId := "idt_12345"

	updateIdentity := &blnkgo.Identity{
		FirstName:    "Jane",
		EmailAddress: "",
		PhoneNumber:  "",
		Category:     "",
		Street:       "",
		Country:      "",
		State:        "",
		PostCode:     "",
		City:         "",
	}

	mockClient.On("NewRequest", fmt.Sprintf("identities/%s", identityId), http.MethodPut, updateIdentity).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{}, nil)

	_, httpResp, err := svc.Update(identityId, updateIdentity)
	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Update_NotFound(t *testing.T) {
	mockClient, svc := setupIdentityService()
	identityId := "12345"

	updateIdentity := &blnkgo.Identity{
		FirstName: "Jane",
	}

	mockClient.On("NewRequest", fmt.Sprintf("identities/%s", identityId), http.MethodPut, updateIdentity).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusNotFound}, errors.New("not found"))

	resp, httpResp, err := svc.Update(identityId, updateIdentity)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "not found")
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Update_ServerError(t *testing.T) {
	mockClient, svc := setupIdentityService()
	identityId := "12345"

	updateIdentity := &blnkgo.Identity{
		FirstName: "Jane",
	}

	mockClient.On("NewRequest", fmt.Sprintf("identities/%s", identityId), http.MethodPut, updateIdentity).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusInternalServerError}, errors.New("server error"))

	resp, httpResp, err := svc.Update(identityId, updateIdentity)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusInternalServerError, httpResp.StatusCode)
	assert.Contains(t, err.Error(), "server error")
	mockClient.AssertExpectations(t)
}
func TestIdentityService_Create_EmptyPayload(t *testing.T) {
	mockClient, svc := setupIdentityService()

	emptyIdentity := blnkgo.Identity{
		FirstName:    "Minimal",
		EmailAddress: "minimal@example.com",
		Category:     "customer",
	}
	expectedResponse := &blnkgo.IdentityResponse{
		IdentityId: "idt_minimal",
		Identity:   emptyIdentity,
	}

	mockClient.On("NewRequest", "identities", http.MethodPost, emptyIdentity).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.IdentityResponse)
		*resp = *expectedResponse
	}).Return(&http.Response{}, nil)

	resp, httpResp, err := svc.Create(emptyIdentity)
	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, expectedResponse, resp)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Create_InvalidIdentityType(t *testing.T) {
	_, svc := setupIdentityService()

	resp, httpResp, err := svc.Create(blnkgo.Identity{
		IdentityType: blnkgo.IdentityType("invalid"),
		FirstName:    "Jane",
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Nil(t, httpResp)
}

func TestIdentityService_Filter_Success(t *testing.T) {
	mockClient, svc := setupIdentityService()

	body := blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "email_address", Operator: blnkgo.OpEqual, Value: "john@example.com"},
		},
		Limit: 10,
	}

	expectedResponse := &blnkgo.FilterResponse{
		Data: []blnkgo.IdentityResponse{
			{
				IdentityId: "idt_1",
				Identity: blnkgo.Identity{
					IdentityType: blnkgo.Individual,
					FirstName:  "John",
					LastName:   "Doe",
					EmailAddress: "john@example.com",
					Category:   "customer",
				},
			},
		},
	}

	mockClient.On("NewRequest", "identities/filter", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		result := args.Get(1).(*blnkgo.FilterResponse)
		*result = *expectedResponse
	})

	result, resp, err := svc.Filter(body)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Filter_WithIncludeCount(t *testing.T) {
	mockClient, svc := setupIdentityService()

	body := blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "category", Operator: blnkgo.OpEqual, Value: "customer"},
		},
		IncludeCount: true,
	}

	count := int64(1)
	expectedResponse := &blnkgo.FilterResponse{
		Data: []blnkgo.IdentityResponse{
			{IdentityId: "idt_1"},
		},
		TotalCount: &count,
	}

	mockClient.On("NewRequest", "identities/filter", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil).Run(func(args mock.Arguments) {
		result := args.Get(1).(*blnkgo.FilterResponse)
		*result = *expectedResponse
	})

	result, resp, err := svc.Filter(body)

	assert.NoError(t, err)
	assert.NotNil(t, result.TotalCount)
	assert.Equal(t, int64(1), *result.TotalCount)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Filter_NewRequestError(t *testing.T) {
	mockClient, svc := setupIdentityService()

	body := blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "first_name", Operator: blnkgo.OpEqual, Value: "Jane"},
		},
	}

	mockClient.On("NewRequest", "identities/filter", http.MethodPost, body).Return(nil, fmt.Errorf("request error"))

	result, resp, err := svc.Filter(body)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Filter_ServerError(t *testing.T) {
	mockClient, svc := setupIdentityService()

	body := blnkgo.FilterParams{
		Filters: []blnkgo.Filter{
			{Field: "first_name", Operator: blnkgo.OpEqual, Value: "Jane"},
		},
	}

	expectedResp := &http.Response{StatusCode: http.StatusInternalServerError}
	mockClient.On("NewRequest", "identities/filter", http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(expectedResp, fmt.Errorf("server error"))

	result, resp, err := svc.Filter(body)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_TokenizeField_Success(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_573ebcc9-4da0-4295-82dc-0fb152b56660"
	field := string(blnkgo.TokenizableFieldFirstName)
	path := "identities/" + identityID + "/tokenize/" + field

	mockClient.On("NewRequest", path, http.MethodPost, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.TokenizeFieldResponse)
		*resp = blnkgo.TokenizeFieldResponse{Message: "Field tokenized successfully"}
	})

	tokenized, httpResp, err := svc.TokenizeField(identityID, field)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, "Field tokenized successfully", tokenized.Message)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_TokenizeField_ValidationErrorEmptyID(t *testing.T) {
	mockClient, svc := setupIdentityService()

	_, _, err := svc.TokenizeField("", string(blnkgo.TokenizableFieldFirstName))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identity id is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestIdentityService_TokenizeField_ValidationErrorEmptyField(t *testing.T) {
	mockClient, svc := setupIdentityService()

	_, _, err := svc.TokenizeField("idt_test_123", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field name is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestIdentityService_TokenizeField_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_test_123"
	field := string(blnkgo.TokenizableFieldEmailAddress)
	path := "identities/" + identityID + "/tokenize/" + field

	mockClient.On("NewRequest", path, http.MethodPost, nil).Return(nil, errors.New("failed to create request"))

	tokenized, httpResp, err := svc.TokenizeField(identityID, field)

	assert.Error(t, err)
	assert.Nil(t, tokenized)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}
