package gochimp3

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	campaigns_path        = "/campaigns"
	single_campaign_path  = campaigns_path + "/%s"
	campaign_content_path = single_campaign_path + "/content"

	send_test_path  = single_campaign_path + "/actions/test"
	send_path       = single_campaign_path + "/actions/send"
	schedule_path   = single_campaign_path + "/actions/schedule"
	unschedule_path = single_campaign_path + "/actions/unschedule"

	CAMPAIGN_TYPE_REGULAR   = "regular"
	CAMPAIGN_TYPE_PLAINTEXT = "plaintext"
	CAMPAIGN_TYPE_ABSPLIT   = "absplit" // deprecated by mailchimp
	CAMPAIGN_TYPE_RSS       = "rss"
	CAMPAIGN_TYPE_VARIATE   = "variate"

	CAMPAIGN_SEND_TYPE_HTML      = "html"
	CAMPAIGN_SEND_TYPE_PLAINTEXT = "plaintext"

	CONDITION_MATCH_ANY = "any"
	CONDITION_MATCH_ALL = "all"

	CONDITION_TYPE_INTERESTS = "Interests"

	CONDITION_OP_CONTAINS = "interestcontains"
)

type CampaignQueryParams struct {
	ExtendedQueryParams

	Type             string
	Status           string
	BeforeSendTime   string
	SinceSendTime    string
	BeforeCreateTime string
	SinceCreateTime  string
	ListId           string
	FolderId         string
	SortField        string
	SortDir          string
}

func (q CampaignQueryParams) Params() map[string]string {
	m := q.ExtendedQueryParams.Params()
	m["type"] = q.Type
	m["status"] = q.Status
	m["before_send_time"] = q.BeforeSendTime
	m["since_send_time"] = q.SinceSendTime
	m["before_create_time"] = q.BeforeCreateTime
	m["since_create_time"] = q.SinceCreateTime
	m["list_id"] = q.ListId
	m["folder_id"] = q.FolderId
	m["sort_field"] = q.SortField
	m["sort_dir"] = q.SortDir
	return m
}

type ListOfCampaigns struct {
	baseList
	Campaigns []*CampaignResponse `json:"campaigns"`
}

type CampaignCreationRecipients struct {
	ListId         string                         `json:"list_id"`
	SegmentOptions CampaignCreationSegmentOptions `json:"segment_opts"`
}

type CampaignCreationSegmentOptions struct {
	SavedSegmentId int    `json:"saved_segment_id,omitempty"`
	Match          string `json:"match,omitempty"` // one of CONDITION_MATCH_*

	// this accepts various payloads. See http://developer.mailchimp.com/documentation/mailchimp/reference/campaigns/#create-post_campaigns
	Conditions interface{} `json:"conditions,omitempty"`
}

type InterestsCondition struct {
	ConditionType string   `json:"condition_type"`
	Field         string   `json:"field"`
	Op            string   `json:"op"`
	Value         []string `json:"value"`
}

type StaticSegmentCondition struct {
	ConditionType string `json:"condition_type"`
	Field         string `json:"field"`
	Op            string `json:"op"`
	Value         int    `json:"value"`
}

func NewStaticSegmentCondition(id int, isEqual bool) *StaticSegmentCondition {
	c := &StaticSegmentCondition{
		ConditionType: "StaticSegment",
		Field:         "static_segment",
		Value:         id,
	}

	if isEqual {
		c.Op = "static_is"
	} else {
		c.Op = "static_not"
	}

	return c
}

type CampaignCreationSettings struct {
	SubjectLine     string `json:"subject_line"`
	PreviewText     string `json:"preview_text"`
	Title           string `json:"title"`
	FromName        string `json:"from_name"`
	ReplyTo         string `json:"reply_to"`
	UseConversation bool   `json:"use_conversation"`
	ToName          string `json:"to_name"`
	FolderId        string `json:"folder_id"`
	Authenticate    bool   `json:"authenticate"`
	AutoFooter      bool   `json:"auto_footer"`
	InlineCss       bool   `json:"inline_css"`
	AutoTweet       bool   `json:"auto_tweet"`
	FbComments      bool   `json:"fb_comments"`
	TemplateId      uint   `json:"template_id"`
}

type CampaignCreationRequest struct {
	Type       string                     `json:"type"` // must be one of the CAMPAIGN_TYPE_* consts
	Recipients CampaignCreationRecipients `json:"recipients"`
	Settings   CampaignCreationSettings   `json:"settings"`
	// variate_settings not implemented
	Tracking CampaignTracking `json:"tracking"`
	// rss_opts not implemented
	// social_card not implemented
}

type CampaignResponseRecipients struct {
	ListId         string `json:"list_id"`
	ListName       string `json:"list_name"`
	SegmentText    string `json:"segment_text"`
	RecipientCount int64  `json:"recipient_count"`
}

type CampaignResponseSettings struct {
	SubjectLine     string `json:"subject_line"`
	PreviewText     string `json:"preview_text"`
	Title           string `json:"title"`
	FromName        string `json:"from_name"`
	ReplyTo         string `json:"reply_to"`
	UseConversation bool   `json:"use_conversation"`
	ToName          string `json:"to_name"`
	FolderId        string `json:"folder_id"`
	Authenticate    bool   `json:"authenticate"`
	AutoFooter      bool   `json:"auto_footer"`
	InlineCss       bool   `json:"inline_css"`
	AutoTweet       bool   `json:"auto_tweet"`
	FbComments      bool   `json:"fb_comments"`
	Timewarp        bool   `json:"timewarp"`
	TemplateId      uint   `json:"template_id"`
	DragAndDrop     bool   `json:"drag_and_drop"`
}

type CampaignTracking struct {
	Opens           bool   `json:"opens"`
	HtmlClicks      bool   `json:"html_clicks"`
	TextClicks      bool   `json:"text_clicks"`
	GoalTracking    bool   `json:"goal_tracking"`
	Ecomm360        bool   `json:"ecomm360"`
	GoogleAnalytics string `json:"google_analytics"`
	Clicktale       string `json:"clicktale"`
}

type CampaignEcommerce struct {
	TotalOrders  int64 `json:"total_orders"`
	TotalSpent   int64 `json:"total_spent"`
	TotalRevenue int64 `json:"total_revenue"`
}

type CampaignReportSummary struct {
	Opens            int64             `json:"opens"`
	UniqueOpens      int64             `json:"unique_opens"`
	OpenRate         float64           `json:"open_rate"`
	Clicks           int64             `json:"clicks"`
	SubscriberClicks int64             `json:"subscriber_clicks"`
	ClickRate        float64           `json:"click_rate"`
	Ecommerce        CampaignEcommerce `json:"ecommerce"`
}

type CampaignDeliveryStatus struct {
	Enabled   bool `json:"enabled"`
	CanCancel bool `json:"can_cancel"`
	// "delivering", "delivered", "canceling", or "canceled".
	Status         string `json:"status"`
	EmailsSent     int64  `json:"emails_sent"`
	EmailsCanceled int64  `json:"emails_canceled"`
}

type CampaignResponse struct {
	withLinks

	ID                string                     `json:"id"`
	WebID             int64                      `json:"web_id"`
	Type              string                     `json:"type"`
	CreateTime        string                     `json:"create_time"`
	ArchiveUrl        string                     `json:"archive_url"`
	LongArchiveUrl    string                     `json:"long_archive_url"`
	Status            string                     `json:"status"`
	EmailsSent        int64                      `json:"emails_sent"`
	SendTime          string                     `json:"send_time"`
	ContentType       string                     `json:"content_type"`
	NeedsBlockRefresh bool                       `json:"needs_block_refresh"`
	Recipients        CampaignResponseRecipients `json:"recipients"`
	Settings          CampaignResponseSettings   `json:"settings"`
	Tracking          CampaignTracking           `json:"tracking"`
	ReportSummary     CampaignReportSummary      `json:"report_summary"`
	DeliveryStatus    CampaignDeliveryStatus     `json:"delivery_status"`

	api *API
}

func (campaign CampaignResponse) CanMakeRequest() error {
	if campaign.ID == "" {
		return errors.New("No ID provided on campaign")
	}

	return nil
}

func (api *API) GetCampaigns(ctx context.Context, params *CampaignQueryParams) (*ListOfCampaigns, error) {
	response := new(ListOfCampaigns)

	err := api.request(ctx, "GET", campaigns_path, params, nil, response)
	if err != nil {
		return nil, err
	}

	for _, l := range response.Campaigns {
		l.api = api
	}

	return response, nil
}

func (api *API) GetCampaign(ctx context.Context, id string, params *BasicQueryParams) (*CampaignResponse, error) {
	endpoint := fmt.Sprintf(single_campaign_path, id)

	response := new(CampaignResponse)
	response.api = api

	return response, api.request(ctx, "GET", endpoint, params, nil, response)
}

func (api *API) CreateCampaign(ctx context.Context, body *CampaignCreationRequest) (*CampaignResponse, error) {
	response := new(CampaignResponse)
	response.api = api
	return response, api.request(ctx, "POST", campaigns_path, nil, body, response)
}

func (api *API) UpdateCampaign(ctx context.Context, id string, body *CampaignCreationRequest) (*CampaignResponse, error) {
	endpoint := fmt.Sprintf(single_campaign_path, id)

	response := new(CampaignResponse)
	response.api = api

	return response, api.request(ctx, "PATCH", endpoint, nil, body, response)
}

func (api *API) DeleteCampaign(ctx context.Context, id string) (bool, error) {
	endpoint := fmt.Sprintf(single_campaign_path, id)
	return api.requestOk(ctx, "DELETE", endpoint)
}

// ------------------------------------------------------------------------------------------------
// Actions
// ------------------------------------------------------------------------------------------------

type TestEmailRequest struct {
	TestEmails []string `json:"test_emails"`
	SendType   string   `json:"send_type"` // one of the CAMPAIGN_SEND_TYPE_* constants
}

type SendCampaignRequest struct {
	CampaignId string `json:"campaign_id"`
}

type ScheduleCampaignRequest struct {
	// The UTC date and time to schedule the campaign for delivery in ISO 8601 format. Campaigns may only be scheduled to send on the quarter-hour (:00, :15, :30, :45).
	ScheduleTime string `json:"schedule_time"`
}

func (api *API) SendTestEmail(ctx context.Context, id string, body *TestEmailRequest) (bool, error) {
	endpoint := fmt.Sprintf(send_test_path, id)
	err := api.request(ctx, "POST", endpoint, nil, body, nil)

	if err != nil {
		return false, err
	}
	return true, nil
}

func (api *API) SendCampaign(ctx context.Context, id string, body *SendCampaignRequest) (bool, error) {
	endpoint := fmt.Sprintf(send_path, id)
	err := api.request(ctx, "POST", endpoint, nil, body, nil)

	if err != nil {
		return false, err
	}
	return true, nil
}

// ScheduleCampaign UTC date and time to schedule the campaign for delivery in ISO 8601 format. Campaigns may only be scheduled to send on the quarter-hour (:00, :15, :30, :45).
func (api *API) ScheduleCampaign(ctx context.Context, id string, scheduleTime *time.Time) (bool, error) {
	body := ScheduleCampaignRequest{
		ScheduleTime: scheduleTime.Format(time.RFC3339),
	}

	endpoint := fmt.Sprintf(schedule_path, id)
	err := api.request(ctx, "POST", endpoint, nil, body, nil)

	if err != nil {
		return false, err
	}
	return true, nil
}

func (api *API) UnscheduleCampaign(ctx context.Context, id string) (bool, error) {
	endpoint := fmt.Sprintf(unschedule_path, id)
	return api.requestOk(ctx, "POST", endpoint)
}

// ------------------------------------------------------------------------------------------------
// Campaign Content Updates
// ------------------------------------------------------------------------------------------------

type CampaignContentTemplateRequest struct {
	ID       uint              `json:"id,omitempty"`
	Sections map[string]string `json:"sections,omitempty"`
}

type CampaignContentUpdateRequest struct {
	PlainText string                          `json:"plain_text"`
	Html      string                          `json:"html"`
	Url       string                          `json:"url"`
	Template  *CampaignContentTemplateRequest `json:"template,omitempty"`
}

type CampaignContentResponse struct {
	withLinks

	PlainText   string `json:"plain_text"`
	Html        string `json:"html"`
	ArchiveHtml string `json:"archive_html"`
	api         *API
}

func (api *API) GetCampaignContent(ctx context.Context, id string, params *BasicQueryParams) (*CampaignContentResponse, error) {
	endpoint := fmt.Sprintf(campaign_content_path, id)
	response := new(CampaignContentResponse)
	response.api = api
	return response, api.request(ctx, "GET", endpoint, nil, params, response)
}

func (api *API) UpdateCampaignContent(ctx context.Context, id string, body *CampaignContentUpdateRequest) (*CampaignContentResponse, error) {
	endpoint := fmt.Sprintf(campaign_content_path, id)
	response := new(CampaignContentResponse)
	response.api = api
	return response, api.request(ctx, "PUT", endpoint, nil, body, response)
}
