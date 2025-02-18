package gochimp3

import "context"

const (
	template_folders_path = "/template-folders"
	// single folder endpoint not implemented
)

type TemplateFolderQueryParams struct {
	ExtendedQueryParams
}

type ListOfTemplateFolders struct {
	baseList
	Folders []TemplateFolder `json:"folders"`
}

type TemplateFolder struct {
	withLinks

	Name  string `json:"name"`
	ID    string `json:"id"`
	Count uint   `json:"count"`

	api *API
}

type TemplateFolderCreationRequest struct {
	Name string `json:"name"`
}

func (api *API) GetTemplateFolders(ctx context.Context, params *TemplateFolderQueryParams) (*ListOfTemplateFolders, error) {
	response := new(ListOfTemplateFolders)

	err := api.request(ctx, "GET", template_folders_path, params, nil, response)
	if err != nil {
		return nil, err
	}

	for _, l := range response.Folders {
		l.api = api
	}

	return response, nil
}

func (api *API) CreateTemplateFolder(ctx context.Context, body *TemplateFolderCreationRequest) (*TemplateFolder, error) {
	response := new(TemplateFolder)
	response.api = api
	return response, api.request(ctx, "POST", template_folders_path, nil, body, response)
}
