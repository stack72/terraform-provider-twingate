package twingate

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestClientConnectorCreateTokensOK(t *testing.T) {
	// response JSON
	createTokensOkJson := `{
		"data": {
			"connectorGenerateTokens": {
				"connectorTokens": {
					"accessToken": "token1",
					"refreshToken": "token2"
				},
				"ok": true,
				"error": null
			}
		}
	}`

	client := newHTTPMockClient()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createTokensOkJson))
	connector := &Connector{
		ID: "test",
	}
	err := client.generateConnectorTokens(context.Background(), connector)

	assert.Nil(t, err)
	assert.EqualValues(t, "token1", connector.ConnectorTokens.AccessToken)
	assert.EqualValues(t, "token2", connector.ConnectorTokens.RefreshToken)
}

func TestClientConnectorTokensVerifyOK(t *testing.T) {
	// response JSON
	verifyTokensOkJson := `{}`

	client := newHTTPMockClient()
	defer httpmock.DeactivateAndReset()

	accessToken := "test1"
	refreshToken := "test2"

	httpmock.RegisterResponder("POST", client.APIServerURL+"/access_node/refresh",
		func(req *http.Request) (*http.Response, error) {
			header := req.Header.Get("Authorization")
			assert.Contains(t, header, accessToken)
			return httpmock.NewStringResponse(200, verifyTokensOkJson), nil
		})

	err := client.verifyConnectorTokens(context.Background(), refreshToken, accessToken)

	assert.Nil(t, err)
}

func TestClientConnectorTokensVerify401Error(t *testing.T) {
	// response JSON
	verifyTokensOkJson := `{}`

	client := newHTTPMockClient()
	defer httpmock.DeactivateAndReset()

	accessToken := "test1"
	refreshToken := "test2"

	apiURL := client.APIServerURL + "/access_node/refresh"
	httpmock.RegisterResponder("POST", apiURL,
		func(req *http.Request) (*http.Response, error) {
			header := req.Header.Get("Authorization")
			assert.Contains(t, header, accessToken)
			return httpmock.NewStringResponse(401, verifyTokensOkJson), nil
		})

	err := client.verifyConnectorTokens(context.Background(), refreshToken, accessToken)

	assert.EqualError(t, err, "failed to verify connector tokens: request https://test.twindev.com/api/v1/access_node/refresh failed, status 401, body {}")
}

func TestClientConnectorTokensVerifyRequestError(t *testing.T) {
	client := newHTTPMockClient()

	accessToken := "test1"
	refreshToken := "test2"

	defer httpmock.DeactivateAndReset()
	apiURL := client.APIServerURL + "/access_node/refresh"
	httpmock.RegisterResponder("POST", apiURL,
		func(req *http.Request) (*http.Response, error) {
			header := req.Header.Get("Authorization")
			assert.Contains(t, header, accessToken)
			return nil, errors.New("error")
		})

	err := client.verifyConnectorTokens(context.Background(), refreshToken, accessToken)
	assert.EqualError(t, err, "failed to verify connector tokens: can't execute http request: Post \"https://test.twindev.com/api/v1/access_node/refresh\": error")
}

func TestClientConnectorCreateTokensError(t *testing.T) {
	// response JSON
	createTokensOkJson := `{
	  "data": {
		"connectorGenerateTokens": {
		  "ok": false,
		  "error": "error_1"
		}
	  }
	}`

	client := newHTTPMockClient()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, createTokensOkJson))
	connector := &Connector{
		ID: "test-id",
	}
	err := client.generateConnectorTokens(context.Background(), connector)

	assert.EqualError(t, err, "failed to generate connector tokens with id test-id: error_1")
}

func TestClientConnectorTokensCreateRequestError(t *testing.T) {
	client := newHTTPMockClient()

	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewErrorResponder(errors.New("error_1")))
	connector := &Connector{
		ID: "test",
	}

	err := client.generateConnectorTokens(context.Background(), connector)
	assert.EqualError(t, err, "failed to generate connector tokens: Post \"https://test.twindev.com/api/graphql/\": error_1")
}
