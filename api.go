package gochimp3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"regexp"
	"time"
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
	CreateBatchOperation(ctx context.Context, body *BatchOperationCreationRequest) (*BatchOperationResponse, error)
	CreateCampaign(ctx context.Context, body *CampaignCreationRequest) (*CampaignResponse, error)
	CreateCampaignFolder(ctx context.Context, body *CampaignFolderCreationRequest) (*CampaignFolder, error)
	CreateList(ctx context.Context, body *ListCreationRequest) (*ListResponse, error)
	CreateTemplate(ctx context.Context, body *TemplateCreationRequest) (*TemplateResponse, error)
	CreateTemplateFolder(ctx context.Context, body *TemplateFolderCreationRequest) (*TemplateFolder, error)
	DeleteCampaign(ctx context.Context, id string) (bool, error)
	DeleteList(ctx context.Context, id string) (bool, error)
	DeleteTemplate(ctx context.Context, id string) (bool, error)
	GetBatchOperation(ctx context.Context, id string, params *BasicQueryParams) (*BatchOperationResponse, error)
	GetBatchOperations(ctx context.Context, params *ListQueryParams) (*ListOfBatchOperations, error)
	GetCampaign(ctx context.Context, id string, params *BasicQueryParams) (*CampaignResponse, error)
	GetCampaignContent(ctx context.Context, id string, params *BasicQueryParams) (*CampaignContentResponse, error)
	GetCampaignFolders(ctx context.Context, params *CampaignFolderQueryParams) (*ListOfCampaignFolders, error)
	GetCampaigns(ctx context.Context, params *CampaignQueryParams) (*ListOfCampaigns, error)
	GetList(ctx context.Context, id string, params *BasicQueryParams) (*ListResponse, error)
	GetLists(ctx context.Context, params *ListQueryParams) (*ListOfLists, error)
	GetRoot(ctx context.Context, params *BasicQueryParams) (*RootResponse, error)
	GetTemplate(ctx context.Context, id string, params *BasicQueryParams) (*TemplateResponse, error)
	GetTemplateDefaultContent(ctx context.Context, id string, params *BasicQueryParams) (*TemplateDefaultContentResponse, error)
	GetTemplateFolders(ctx context.Context, params *TemplateFolderQueryParams) (*ListOfTemplateFolders, error)
	GetTemplates(ctx context.Context, params *TemplateQueryParams) (*ListOfTemplates, error)
	MemberForApiCalls(listId string, email string) *Member
	NewListResponse(id string) *ListResponse
	SendCampaign(ctx context.Context, id string, body *SendCampaignRequest) (bool, error)
	// ScheduleCampaign UTC date and time to schedule the campaign for delivery in ISO 8601 format. Campaigns may only be scheduled to send on the quarter-hour (:00, :15, :30, :45).
	ScheduleCampaign(ctx context.Context, id string, scheduleTime *time.Time) (bool, error)
	UnscheduleCampaign(ctx context.Context, id string) (bool, error)
	SendTestEmail(ctx context.Context, id string, body *TestEmailRequest) (bool, error)
	UpdateCampaign(ctx context.Context, id string, body *CampaignCreationRequest) (*CampaignResponse, error)
	UpdateCampaignContent(ctx context.Context, id string, body *CampaignContentUpdateRequest) (*CampaignContentResponse, error)
	UpdateList(ctx context.Context, id string, body *ListCreationRequest) (*ListResponse, error)
	UpdateTemplate(ctx context.Context, id string, body *TemplateCreationRequest) (*TemplateResponse, error)
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
func (api *API) request(ctx context.Context, method, path string, params QueryParams, body, response interface{}) error {
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

	req, err := http.NewRequestWithContext(ctx, method, requestURL, bodyBytes)
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

	data, err = io.ReadAll(resp.Body)
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

// requestOk Make Request ignoring body and return true if HTTP status code is 2xx.
func (api *API) requestOk(ctx context.Context, method, path string) (bool, error) {
	err := api.request(ctx, method, path, nil, nil, nil)
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
