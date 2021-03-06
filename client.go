package hubspot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

// ClientConfig object used for client creation
type ClientConfig struct {
	APIHost    string
	APIKey     string
	OAuthToken string
}

// NewClientConfig constructs a ClientConfig object with the environment variables set as default
func NewClientConfig() ClientConfig {
	return ClientConfig{
		APIHost:    apiHost,
		APIKey:     apiKey,
		OAuthToken: oauthToken,
	}
}

// Client object
type Client struct {
	config ClientConfig
}

// NewClient constructor
func NewClient(config ClientConfig) Client {
	return Client{
		config: config,
	}
}

// addAPIKey adds HUBSPOT_API_KEY param to a given URL.
func (c Client) addAPIKey(u string) (string, error) {
	if c.config.APIKey != "" {
		uri, err := url.Parse(u)
		if err != nil {
			return u, err
		}
		q := uri.Query()
		q.Set("hapikey", c.config.APIKey)
		uri.RawQuery = q.Encode()
		u = uri.String()
	}

	return u, nil
}

// Request executes any HubSpot API method using the current client configuration
func (c Client) Request(method, endpoint string, data, response interface{}) error {
	// Construct endpoint URL
	u, err := url.Parse(c.config.APIHost)
	if err != nil {
		return fmt.Errorf("hubspot.Client.Request(): url.Parse(): %v", err)
	}
	u.Path = path.Join(u.Path, endpoint)

	// API Key authentication
	uri := u.String()
	if c.config.APIKey != "" {
		uri, err = c.addAPIKey(uri)
		if err != nil {
			return fmt.Errorf("hubspot.Client.Request(): c.addAPIKey(): %v", err)
		}
	}

	// Init request object
	var req *http.Request

	// Send data?
	if data != nil {
		// Encode data to JSON
		dataEncoded, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("hubspot.Client.Request(): json.Marshal(): %v", err)
		}
		buf := bytes.NewBuffer(dataEncoded)

		// Create request
		req, err = http.NewRequest(method, uri, buf)
	} else {
		// Create no-data request
		req, err = http.NewRequest(method, uri, nil)
	}
	if err != nil {
		return fmt.Errorf("hubspot.Client.Request(): http.NewRequest(): %v", err)
	}

	// OAuth authentication
	if c.config.APIKey == "" && c.config.OAuthToken != "" {
		req.Header.Add("Authorization", "Bearer "+c.config.OAuthToken)
	}

	// Headers
	req.Header.Add("Content-Type", "application/json")

	// Execute and read response body
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("hubspot.Client.Request(): httpClient.Do(): %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("hubspot.Client.Request(): ioutil.ReadAll(): %v", err)
	}

	// Get data?
	if response != nil {
		err = json.Unmarshal(body, &response)
		if err != nil {
			return fmt.Errorf("hubspot.Client.Request(): json.Unmarshal(): %v \n%s", err, string(body))
		}
	}

	// Return HTTP errors
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("HubSpot API error: %d - %s \n%s", resp.StatusCode, resp.Status, string(body))
	}

	// Done!
	return nil
}
