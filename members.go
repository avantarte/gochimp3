package gochimp3

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	members_path       = "/lists/%s/members"
	single_member_path = members_path + "/%s"

	member_activity_path = single_member_path + "/activity"
	member_goals_path    = single_member_path + "/goals"

	member_notes_path       = single_member_path + "/notes"
	single_member_note_path = member_notes_path + "/%s"

	member_tags_path       = single_member_path + "/tags"
	single_member_tag_path = member_tags_path + "/%s"

	delete_permanent_path = single_member_path + "/actions/delete-permanent"
)

type ListOfMembers struct {
	baseList

	ListID  string   `json:"list_id"`
	Members []Member `json:"members"`
}

type MemberResponse struct {
	EmailAddress    string                 `json:"email_address"`
	EmailType       string                 `json:"email_type,omitempty"`
	Status          string                 `json:"status"`
	StatusIfNew     string                 `json:"status_if_new,omitempty"`
	MergeFields     map[string]interface{} `json:"merge_fields,omitempty"`
	Interests       map[string]bool        `json:"interests,omitempty"`
	Language        string                 `json:"language"`
	VIP             bool                   `json:"vip"`
	Location        *MemberLocation        `json:"location,omitempty"`
	IPOpt           string                 `json:"ip_opt,omitempty"`
	IPSignup        string                 `json:"ip_signup,omitempty"`
	Tags            []MemberTag            `json:"tags,omitempty"`
	TimestampSignup string                 `json:"timestamp_signup,omitempty"`
	TimestampOpt    string                 `json:"timestamp_opt,omitempty"`
}

type MemberRequest struct {
	EmailAddress         string                 `json:"email_address"`
	EmailType            string                 `json:"email_type,omitempty"`
	Status               string                 `json:"status"`
	StatusIfNew          string                 `json:"status_if_new,omitempty"`
	MergeFields          map[string]interface{} `json:"merge_fields,omitempty"`
	Interests            map[string]bool        `json:"interests,omitempty"`
	Language             string                 `json:"language"`
	VIP                  bool                   `json:"vip"`
	Location             *MemberLocation        `json:"location,omitempty"`
	MarketingPermissions *MarketingPermissions  `json:"marketing_permissions,omitempty"`
	IPOpt                string                 `json:"ip_opt,omitempty"`
	IPSignup             string                 `json:"ip_signup,omitempty"`
	Tags                 []string               `json:"tags,omitempty"`
	TimestampSignup      string                 `json:"timestamp_signup,omitempty"`
	TimestampOpt         string                 `json:"timestamp_opt,omitempty"`
}

type Member struct {
	MemberResponse

	ID            string          `json:"id"`
	ListID        string          `json:"list_id"`
	UniqueEmailID string          `json:"unique_email_id"`
	EmailType     string          `json:"email_type"`
	Stats         MemberStats     `json:"stats"`
	MemberRating  int             `json:"member_rating"`
	LastChanged   string          `json:"last_changed"`
	EmailClient   string          `json:"email_client"`
	LastNote      MemberNoteShort `json:"last_note"`

	api *API
}

func (mem *Member) CanMakeRequest() error {
	if mem.ListID == "" {
		return errors.New("No ListID provided")
	}

	if mem.ID == "" {
		return errors.New("No ID provided")
	}

	return nil
}

func (mem *Member) WithApi(api *API) *Member {
	mem.api = api
	return mem
}

func (mem *Member) SetIdByMail(email string) *Member {
	id, err := EmailToMemberID(email)
	if err != nil {
		panic(err)
	}
	mem.ID = id
	return mem
}

// The Member struct returned by this should be all you need to start calling APIs for this member,
// meaning it should pass CanMakeRequest.
func (api *API) MemberForApiCalls(listId, email string) *Member {
	return (&Member{ListID: listId}).WithApi(api).SetIdByMail(email)
}

type MemberStats struct {
	AvgOpenRate  float64 `json:"avg_open_rate"`
	AvgClickRate float64 `json:"avg_click_rate"`
}

type MemberLocation struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	GMTOffset   int     `json:"gmtoff"`
	DSTOffset   int     `json:"dstoff"`
	CountryCode string  `json:"country_code"`
	Timezone    string  `json:"timezone"`
}

type MarketingPermissions []MarketingPermission

type MarketingPermission struct {
	MarketingPermissionID string `json:"marketing_permission_id"`
	Text                  string `json:"text"`
	Enabled               bool   `json:"enabled"`
}

type MemberNoteShort struct {
	ID        int    `json:"note_id"`
	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
	Note      string `json:"note"`
}

type MemberTag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (list *ListResponse) GetMembers(ctx context.Context, params *InterestCategoriesQueryParams) (*ListOfMembers, error) {
	if err := list.CanMakeRequest(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(members_path, list.ID)
	response := new(ListOfMembers)

	err := list.api.request(ctx, "GET", endpoint, params, nil, response)
	if err != nil {
		return nil, err
	}

	for _, m := range response.Members {
		m.api = list.api
	}

	return response, nil
}

func (api *API) ListGetMembers(ctx context.Context, listID string, params *InterestCategoriesQueryParams) (*ListOfMembers, error) {
	return api.NewListResponse(listID).GetMembers(ctx, params)
}

func (list *ListResponse) GetMember(ctx context.Context, id string, params *BasicQueryParams) (*Member, error) {
	if err := list.CanMakeRequest(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(single_member_path, list.ID, id)
	response := new(Member)
	response.api = list.api

	return response, list.api.request(ctx, "GET", endpoint, params, nil, response)
}

func (list *ListResponse) CreateMember(ctx context.Context, body *MemberRequest) (*Member, error) {
	if err := list.CanMakeRequest(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(members_path, list.ID)
	response := new(Member)
	response.api = list.api

	return response, list.api.request(ctx, "POST", endpoint, nil, body, response)
}

func (list *ListResponse) UpdateMember(ctx context.Context, id string, body *MemberRequest) (*Member, error) {
	if err := list.CanMakeRequest(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(single_member_path, list.ID, id)
	response := new(Member)
	response.api = list.api

	return response, list.api.request(ctx, "PATCH", endpoint, nil, body, response)
}

func (list *ListResponse) AddOrUpdateMember(ctx context.Context, id string, body *MemberRequest) (*Member, error) {
	if err := list.CanMakeRequest(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(single_member_path, list.ID, id)
	response := new(Member)
	response.api = list.api

	return response, list.api.request(ctx, "PUT", endpoint, nil, body, response)
}

func (api *API) ListAddOrUpdateMember(ctx context.Context, listID, memberID string, body *MemberRequest) (*Member, error) {
	if memberID == "" {
		if body.EmailAddress == "" {
			return nil, errors.New("email address or memberID is required")
		}

		var err error

		memberID, err = EmailToMemberID(body.EmailAddress)
		if err != nil {
			return nil, err
		}
	}

	return api.NewListResponse(listID).AddOrUpdateMember(ctx, memberID, body)
}

func (list *ListResponse) DeleteMember(ctx context.Context, id string) (bool, error) {
	if err := list.CanMakeRequest(); err != nil {
		return false, err
	}

	endpoint := fmt.Sprintf(single_member_path, list.ID, id)
	return list.api.requestOk(ctx, "DELETE", endpoint)
}

func (list *ListResponse) DeleteMemberPermanent(ctx context.Context, id string) (bool, error) {
	if err := list.CanMakeRequest(); err != nil {
		return false, err
	}

	endpoint := fmt.Sprintf(delete_permanent_path, list.ID, id)
	return list.api.requestOk(ctx, "POST", endpoint)
}

// ------------------------------------------------------------------------------------------------
// Activity
// ------------------------------------------------------------------------------------------------

type ListOfMemberActivity struct {
	baseList

	EmailID  string     `json:"email_id"`
	ListID   string     `json:"list_id"`
	Activity []Activity `json:"activity"`
}

type MemberActivity struct {
	Action         string `json:"action"`
	Timestamp      string `json:"timestamp"`
	URL            string `json:"url"`
	Type           string `json:"type"`
	CampaignID     string `json:"campaign_id"`
	Title          string `json:"title"`
	ParentCampaign string `json:"parent_campaign"`
}

func (mem *Member) GetActivity(ctx context.Context, params *BasicQueryParams) (*ListOfMemberActivity, error) {
	if err := mem.CanMakeRequest(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(member_activity_path, mem.ListID, mem.ID)
	response := new(ListOfMemberActivity)

	return response, mem.api.request(ctx, "GET", endpoint, params, nil, response)
}

// ------------------------------------------------------------------------------------------------
// Goals
// ------------------------------------------------------------------------------------------------

type ListOfMemberGoals struct {
	baseList

	EmailID string       `json:"email_id"`
	ListID  string       `json:"list_id"`
	Goals   []MemberGoal `json:"goals"`
}

type MemberGoal struct {
	ID            int    `json:"goal_id"`
	Event         string `json:"event"`
	LastVisitedAt string `json:"last_visited_at"`
	Data          string `json:"data"`
}

func (mem *Member) GetGoals(ctx context.Context, params *BasicQueryParams) (*ListOfMemberGoals, error) {
	if err := mem.CanMakeRequest(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(member_goals_path, mem.ListID, mem.ID)
	response := new(ListOfMemberGoals)

	return response, mem.api.request(ctx, "GET", endpoint, params, nil, response)
}

// ------------------------------------------------------------------------------------------------
// NOTES
// ------------------------------------------------------------------------------------------------

type ListOfMemberNotes struct {
	baseList

	EmailID string           `json:"email_id"`
	ListID  string           `json:"list_id"`
	Notes   []MemberNoteLong `json:"notes"`
}

type MemberNoteLong struct {
	ID        int    `json:"id"`
	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
	UpdatedAt string `json:"updated_at"`
	Note      string `json:"note"`
	ListID    string `json:"list_id"`
	EmailID   string `json:"email_id"`

	withLinks
}

func (mem *Member) GetNotes(ctx context.Context, params *ExtendedQueryParams) (*ListOfMemberNotes, error) {
	if err := mem.CanMakeRequest(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(member_notes_path, mem.ListID, mem.ID)
	response := new(ListOfMemberNotes)

	return response, mem.api.request(ctx, "GET", endpoint, params, nil, response)
}

func (mem *Member) CreateNote(ctx context.Context, msg string) (*MemberNoteLong, error) {
	if err := mem.CanMakeRequest(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(member_notes_path, mem.ListID, mem.ID)
	response := new(MemberNoteLong)

	body := struct{ Note string }{
		Note: msg,
	}

	return response, mem.api.request(ctx, "POST", endpoint, nil, &body, response)
}

func (mem *Member) UpdateNote(ctx context.Context, id, msg string) (*MemberNoteLong, error) {
	if err := mem.CanMakeRequest(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(single_member_note_path, mem.ListID, mem.ID, id)
	response := new(MemberNoteLong)

	body := struct{ Note string }{
		Note: msg,
	}

	return response, mem.api.request(ctx, "PATCH", endpoint, nil, &body, response)
}

func (mem *Member) GetNote(ctx context.Context, id string, params *BasicQueryParams) (*MemberNoteLong, error) {
	if err := mem.CanMakeRequest(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(single_member_note_path, mem.ListID, mem.ID, id)
	response := new(MemberNoteLong)

	return response, mem.api.request(ctx, "GET", endpoint, params, nil, response)
}

func (mem *Member) DeleteNote(ctx context.Context, id string) (bool, error) {
	if err := mem.CanMakeRequest(); err != nil {
		return false, err
	}

	endpoint := fmt.Sprintf(single_member_note_path, mem.ListID, mem.ID, id)
	return mem.api.requestOk(ctx, "DELETE", endpoint)
}

// ------------------------------------------------------------------------------------------------
// TAGS
// ------------------------------------------------------------------------------------------------

type ListOfMemberTags struct {
	baseList

	Tags []MemberTagLong `json:"tags"`
}

type UpdateMemberTag struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type MemberTagLong struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	DataAdded *string `json:"date_added,omitempty"`
	Status    string  `json:"status,omitempty"`

	withLinks
}

func (mem *Member) GetTags(ctx context.Context, params *ExtendedQueryParams) (*ListOfMemberTags, error) {
	if err := mem.CanMakeRequest(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(member_tags_path, mem.ListID, mem.ID)
	response := new(ListOfMemberTags)

	return response, mem.api.request(ctx, "GET", endpoint, params, nil, response)
}

func (mem *Member) UpdateTags(ctx context.Context, tags []UpdateMemberTag) (*ListOfMemberTags, error) {
	if err := mem.CanMakeRequest(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(member_tags_path, mem.ListID, mem.ID)
	response := new(ListOfMemberTags)

	body := struct {
		Tags []UpdateMemberTag `json:"tags,omitempty"`
	}{
		Tags: tags,
	}

	return response, mem.api.request(ctx, "POST", endpoint, nil, &body, response)
}

// EmailToMemberID converts given email address to subscriber_hash to be used as ID in Member API.
func EmailToMemberID(email string) (string, error) {
	h := md5.New()

	_, err := io.WriteString(h, strings.ToLower(email))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
