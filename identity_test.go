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

func TestIdentityService_Tokenize_Success(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_573ebcc9-4da0-4295-82dc-0fb152b56660"
	body := blnkgo.TokenizeRequest{
		Fields: []blnkgo.TokenizableIdentityField{
			blnkgo.TokenizableFieldFirstName,
			blnkgo.TokenizableFieldLastName,
		},
	}
	path := "identities/" + identityID + "/tokenize"

	mockClient.On("NewRequest", path, http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.TokenizeResponse)
		*resp = blnkgo.TokenizeResponse{Message: "Fields tokenized successfully"}
	})

	tokenized, httpResp, err := svc.Tokenize(identityID, body)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, "Fields tokenized successfully", tokenized.Message)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Tokenize_ValidationErrorEmptyID(t *testing.T) {
	mockClient, svc := setupIdentityService()

	body := blnkgo.TokenizeRequest{Fields: []blnkgo.TokenizableIdentityField{blnkgo.TokenizableFieldFirstName}}
	_, _, err := svc.Tokenize("", body)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identity id is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestIdentityService_Tokenize_ValidationErrorEmptyFields(t *testing.T) {
	mockClient, svc := setupIdentityService()

	_, _, err := svc.Tokenize("idt_test_123", blnkgo.TokenizeRequest{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one field must be specified")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestIdentityService_Tokenize_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_test_123"
	body := blnkgo.TokenizeRequest{Fields: []blnkgo.TokenizableIdentityField{blnkgo.TokenizableFieldEmailAddress}}
	path := "identities/" + identityID + "/tokenize"

	mockClient.On("NewRequest", path, http.MethodPost, body).Return(nil, errors.New("failed to create request"))

	tokenized, httpResp, err := svc.Tokenize(identityID, body)

	assert.Error(t, err)
	assert.Nil(t, tokenized)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_GetTokenizedFields_Success(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_573ebcc9-4da0-4295-82dc-0fb152b56660"
	path := "identities/" + identityID + "/tokenized-fields"

	mockClient.On("NewRequest", path, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.GetTokenizedFieldsResponse)
		*resp = blnkgo.GetTokenizedFieldsResponse{
			TokenizedFields: []blnkgo.TokenizableIdentityField{
				blnkgo.TokenizableFieldFirstName,
				blnkgo.TokenizableFieldEmailAddress,
			},
		}
	})

	fields, httpResp, err := svc.GetTokenizedFields(identityID)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Len(t, fields.TokenizedFields, 2)
	assert.Equal(t, blnkgo.TokenizableFieldFirstName, fields.TokenizedFields[0])
	mockClient.AssertExpectations(t)
}

func TestIdentityService_GetTokenizedFields_EmptyList(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_test_123"
	path := "identities/" + identityID + "/tokenized-fields"

	mockClient.On("NewRequest", path, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.GetTokenizedFieldsResponse)
		*resp = blnkgo.GetTokenizedFieldsResponse{TokenizedFields: []blnkgo.TokenizableIdentityField{}}
	})

	fields, httpResp, err := svc.GetTokenizedFields(identityID)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Empty(t, fields.TokenizedFields)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_GetTokenizedFields_ValidationError(t *testing.T) {
	mockClient, svc := setupIdentityService()

	_, _, err := svc.GetTokenizedFields("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identity id is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestIdentityService_GetTokenizedFields_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_test_123"
	path := "identities/" + identityID + "/tokenized-fields"

	mockClient.On("NewRequest", path, http.MethodGet, nil).Return(nil, errors.New("failed to create request"))

	fields, httpResp, err := svc.GetTokenizedFields(identityID)

	assert.Error(t, err)
	assert.Nil(t, fields)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_DetokenizeField_Success(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_573ebcc9-4da0-4295-82dc-0fb152b56660"
	field := string(blnkgo.TokenizableFieldEmailAddress)
	path := "identities/" + identityID + "/detokenize/" + field

	mockClient.On("NewRequest", path, http.MethodGet, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.DetokenizeFieldResponse)
		*resp = blnkgo.DetokenizeFieldResponse{
			Field: "EmailAddress",
			Value: "alice.smith@example.com",
		}
	})

	detokenized, httpResp, err := svc.DetokenizeField(identityID, field)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, "EmailAddress", detokenized.Field)
	assert.Equal(t, "alice.smith@example.com", detokenized.Value)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_DetokenizeField_ValidationErrorEmptyID(t *testing.T) {
	mockClient, svc := setupIdentityService()

	_, _, err := svc.DetokenizeField("", string(blnkgo.TokenizableFieldFirstName))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identity id is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestIdentityService_DetokenizeField_ValidationErrorEmptyField(t *testing.T) {
	mockClient, svc := setupIdentityService()

	_, _, err := svc.DetokenizeField("idt_test_123", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field name is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestIdentityService_DetokenizeField_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_test_123"
	field := string(blnkgo.TokenizableFieldPhoneNumber)
	path := "identities/" + identityID + "/detokenize/" + field

	mockClient.On("NewRequest", path, http.MethodGet, nil).Return(nil, errors.New("failed to create request"))

	detokenized, httpResp, err := svc.DetokenizeField(identityID, field)

	assert.Error(t, err)
	assert.Nil(t, detokenized)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Detokenize_Success(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_573ebcc9-4da0-4295-82dc-0fb152b56660"
	body := blnkgo.DetokenizeRequest{
		Fields: []blnkgo.TokenizableIdentityField{
			blnkgo.TokenizableFieldFirstName,
			blnkgo.TokenizableFieldEmailAddress,
		},
	}
	path := "identities/" + identityID + "/detokenize"

	mockClient.On("NewRequest", path, http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.DetokenizeResponse)
		*resp = blnkgo.DetokenizeResponse{
			Fields: map[string]string{
				"FirstName":    "Jane",
				"EmailAddress": "jane@example.com",
			},
		}
	})

	detokenized, httpResp, err := svc.Detokenize(identityID, body)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, "Jane", detokenized.Fields["FirstName"])
	assert.Equal(t, "jane@example.com", detokenized.Fields["EmailAddress"])
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Detokenize_EmptyFields(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_test_123"
	body := blnkgo.DetokenizeRequest{Fields: []blnkgo.TokenizableIdentityField{}}
	path := "identities/" + identityID + "/detokenize"

	mockClient.On("NewRequest", path, http.MethodPost, body).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.DetokenizeResponse)
		*resp = blnkgo.DetokenizeResponse{
			Fields: map[string]string{
				"FirstName":    "Jane",
				"EmailAddress": "jane@example.com",
			},
		}
	})

	detokenized, httpResp, err := svc.Detokenize(identityID, body)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Len(t, detokenized.Fields, 2)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Detokenize_ValidationErrorEmptyID(t *testing.T) {
	mockClient, svc := setupIdentityService()

	body := blnkgo.DetokenizeRequest{Fields: []blnkgo.TokenizableIdentityField{blnkgo.TokenizableFieldFirstName}}
	_, _, err := svc.Detokenize("", body)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identity id is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestIdentityService_Detokenize_ValidationErrorNilFields(t *testing.T) {
	mockClient, svc := setupIdentityService()

	_, _, err := svc.Detokenize("idt_test_123", blnkgo.DetokenizeRequest{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fields must be an array")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestIdentityService_Detokenize_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_test_123"
	body := blnkgo.DetokenizeRequest{Fields: []blnkgo.TokenizableIdentityField{blnkgo.TokenizableFieldPhoneNumber}}
	path := "identities/" + identityID + "/detokenize"

	mockClient.On("NewRequest", path, http.MethodPost, body).Return(nil, errors.New("failed to create request"))

	detokenized, httpResp, err := svc.Detokenize(identityID, body)

	assert.Error(t, err)
	assert.Nil(t, detokenized)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Delete_Success(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_573ebcc9-4da0-4295-82dc-0fb152b56660"
	path := "identities/" + identityID

	mockClient.On("NewRequest", path, http.MethodDelete, nil).Return(&http.Request{}, nil)
	mockClient.On("CallWithRetry", mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Run(func(args mock.Arguments) {
		resp := args.Get(1).(*blnkgo.DeleteIdentityResponse)
		*resp = blnkgo.DeleteIdentityResponse{Message: "Identity deleted successfully"}
	})

	deleted, httpResp, err := svc.Delete(identityID)

	assert.NoError(t, err)
	assert.NotNil(t, httpResp)
	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, "Identity deleted successfully", deleted.Message)
	mockClient.AssertExpectations(t)
}

func TestIdentityService_Delete_ValidationError(t *testing.T) {
	mockClient, svc := setupIdentityService()

	_, _, err := svc.Delete("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identity id is required")
	mockClient.AssertNotCalled(t, "NewRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestIdentityService_Delete_RequestCreationFailure(t *testing.T) {
	mockClient, svc := setupIdentityService()

	identityID := "idt_test_123"
	path := "identities/" + identityID

	mockClient.On("NewRequest", path, http.MethodDelete, nil).Return(nil, errors.New("failed to create request"))

	deleted, httpResp, err := svc.Delete(identityID)

	assert.Error(t, err)
	assert.Nil(t, deleted)
	assert.Nil(t, httpResp)
	mockClient.AssertExpectations(t)
}
