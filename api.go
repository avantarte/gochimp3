package gochimp3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"regexp"
)

// URIFormat defines the endpoint for a single app
const URIFormat string = "%s.api.mailchimp.com"

// Version the latest API version
const Version string = "/3.0"

// DatacenterRegex defines which datacenter to hit
var DatacenterRegex = regexp.MustCompile("[^-]\\w+$")

// API represents the origin of the API
type API struct {
	Key    string
	Client *http.Client

	User  string
	Debug bool

	endpoint string
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Mailchimp
type Mailchimp interface {
	CreateBatchOperation(body *BatchOperationCreationRequest) (*BatchOperationResponse, error)
	CreateCampaign(body *CampaignCreationRequest) (*CampaignResponse, error)
	CreateCampaignFolder(body *CampaignFolderCreationRequest) (*CampaignFolder, error)
	CreateList(body *ListCreationRequest) (*ListResponse, error)
	CreateTemplate(body *TemplateCreationRequest) (*TemplateResponse, error)
	CreateTemplateFolder(body *TemplateFolderCreationRequest) (*TemplateFolder, error)
	DeleteCampaign(id string) (bool, error)
	DeleteList(id string) (bool, error)
	DeleteStore(id string) (bool, error)
	DeleteTemplate(id string) (bool, error)
	GetBatchOperation(id string, params *BasicQueryParams) (*BatchOperationResponse, error)
	GetBatchOperations(params *ListQueryParams) (*ListOfBatchOperations, error)
	GetCampaign(id string, params *BasicQueryParams) (*CampaignResponse, error)
	GetCampaignContent(id string, params *BasicQueryParams) (*CampaignContentResponse, error)
	GetCampaignFolders(params *CampaignFolderQueryParams) (*ListOfCampaignFolders, error)
	GetCampaigns(params *CampaignQueryParams) (*ListOfCampaigns, error)
	GetList(id string, params *BasicQueryParams) (*ListResponse, error)
	GetLists(params *ListQueryParams) (*ListOfLists, error)
	GetRoot(params *BasicQueryParams) (*RootResponse, error)
	GetTemplate(id string, params *BasicQueryParams) (*TemplateResponse, error)
	GetTemplateDefaultContent(id string, params *BasicQueryParams) (*TemplateDefaultContentResponse, error)
	GetTemplateFolders(params *TemplateFolderQueryParams) (*ListOfTemplateFolders, error)
	GetTemplates(params *TemplateQueryParams) (*ListOfTemplates, error)
	MemberForApiCalls(listId string, email string) *Member
	NewListResponse(id string) *ListResponse
	PauseSending(workflowID string, emailID string) (bool, error)
	PauseSendingAll(id string) (bool, error)
	Request(method string, path string, params QueryParams, body interface{}, response interface{}) error
	RequestOk(method string, path string) (bool, error)
	SendCampaign(id string, body *SendCampaignRequest) (bool, error)
	SendTestEmail(id string, body *TestEmailRequest) (bool, error)
	StartSending(workflowID string, emailID string) (bool, error)
	StartSendingAll(id string) (bool, error)
	UpdateCampaign(id string, body *CampaignCreationRequest) (*CampaignResponse, error)
	UpdateCampaignContent(id string, body *CampaignContentUpdateRequest) (*CampaignContentResponse, error)
	UpdateList(id string, body *ListCreationRequest) (*ListResponse, error)
	UpdateTemplate(id string, body *TemplateCreationRequest) (*TemplateResponse, error)
}

var _ Mailchimp = &API{}

// New creates a API
func New(apiKey string, client *http.Client) *API {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = fmt.Sprintf(URIFormat, DatacenterRegex.FindString(apiKey))
	u.Path = Version

	if client == nil {
		client = http.DefaultClient
	}

	return &API{
		User:     "gochimp3",
		Key:      apiKey,
		endpoint: u.String(),
		Client:   client,
	}
}

// Request will make a call to the actual API.
func (api *API) Request(method, path string, params QueryParams, body, response interface{}) error {
	requestURL := fmt.Sprintf("%s%s", api.endpoint, path)
	if api.Debug {
		log.Printf("Requesting %s: %s\n", method, requestURL)
	}

	var bodyBytes io.Reader
	var err error
	var data []byte
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return err
		}
		bodyBytes = bytes.NewBuffer(data)
		if api.Debug {
			log.Printf("Adding body: %+v\n", body)
		}
	}

	req, err := http.NewRequest(method, requestURL, bodyBytes)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(api.User, api.Key)

	if params != nil && !reflect.ValueOf(params).IsNil() {
		queryParams := req.URL.Query()
		for k, v := range params.Params() {
			if v != "" {
				queryParams.Set(k, v)
			}
		}
		req.URL.RawQuery = queryParams.Encode()

		if api.Debug {
			log.Printf("Adding query params: %q\n", req.URL.Query())
		}
	}

	if api.Debug {
		dump, _ := httputil.DumpRequestOut(req, true)
		log.Printf("%s", string(dump))
	}

	resp, err := api.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if api.Debug {
		dump, _ := httputil.DumpResponse(resp, true)
		log.Printf("%s", string(dump))
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Do not unmarshall response is nil
		if response == nil || reflect.ValueOf(response).IsNil() || len(data) == 0 {
			return nil
		}

		err = json.Unmarshal(data, response)
		if err != nil {
			return err
		}

		return nil
	}

	// This is an API Error
	return parseAPIError(data)
}

// RequestOk Make Request ignoring body and return true if HTTP status code is 2xx.
func (api *API) RequestOk(method, path string) (bool, error) {
	err := api.Request(method, path, nil, nil, nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

func parseAPIError(data []byte) error {
	apiError := new(APIError)
	err := json.Unmarshal(data, apiError)
	if err != nil {
		return err
	}

	return apiError
}
