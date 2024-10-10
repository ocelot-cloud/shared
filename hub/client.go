package hub

import (
	"encoding/json"
	"fmt"
	"github.com/ocelot-cloud/shared/utils"
)

var (
	Port    = "8082"
	RootUrl = "http://localhost:" + Port

	RegistrationPath = "/registration"
	LoginPath        = "/login"
	LogoutPath       = "/logout"
	AuthCheckPath    = "/auth-check"
	WipeDataPath     = "/wipe-data"

	UserPath           = "/user"
	DeleteUserPath     = UserPath + "/delete"
	ChangePasswordPath = UserPath + "/password"

	TagPath       = "/tags"
	TagUploadPath = TagPath + "/upload"
	TagDeletePath = TagPath + "/delete"
	GetTagsPath   = TagPath + "/get-tags"
	DownloadPath  = TagPath + "/" // TODO maybe add "download" to make it clearer?

	AppPath         = "/apps"
	AppCreationPath = AppPath + "/create"
	AppGetListPath  = AppPath + "/get-list"
	AppDeletePath   = AppPath + "/delete"
	SearchAppsPath  = AppPath + "/search"
)

var (
	SampleUser           = "myuser"
	SampleApp            = "myapp"
	SampleTag            = "v0.0.1"
	SampleTagFileContent = "hello"
	SampleEmail          = "testuser@example.com"
	SamplePassword       = "mypassword"
	SampleOrigin         = RootUrl
	SampleForm           = &RegistrationForm{
		SampleUser,
		SamplePassword,
		SampleEmail,
	}
)

type HubClient struct {
	Parent        utils.ComponentClient
	Email         string
	App           string
	Tag           string
	UploadContent []byte
}

func GetRegistrationForm(hub *HubClient) *RegistrationForm {
	return &RegistrationForm{
		User:     hub.Parent.User,
		Password: hub.Parent.Password,
		Email:    hub.Email,
	}
}

func GetHub() *HubClient {
	hub := &HubClient{
		Parent: utils.ComponentClient{
			User:            SampleUser,
			Password:        SamplePassword,
			Origin:          RootUrl,
			SetOriginHeader: true,
			SetCookieHeader: true,
			RootUrl:         RootUrl,
		},

		Email:         SampleEmail,
		App:           SampleApp,
		Tag:           SampleTag,
		UploadContent: []byte(SampleTagFileContent),
	}
	hub.WipeData()
	return hub
}

func (h *HubClient) RegisterUser() error {
	form := GetRegistrationForm(h)
	_, err := h.Parent.DoRequest(RegistrationPath, form, "")
	return err
}

func (h *HubClient) Login() error {
	creds := LoginCredentials{
		User:     h.Parent.User,
		Password: h.Parent.Password,
		Origin:   h.Parent.Origin,
	}

	resp, err := h.Parent.DoRequestWithFullResponse(LoginPath, creds, "")
	if err != nil {
		return err
	}

	cookies := resp.Cookies()
	if len(cookies) != 1 {
		return fmt.Errorf("Expected 1 cookie, got %d", len(cookies))
	}
	h.Parent.Cookie = cookies[0]
	return nil
}

func (h *HubClient) DeleteUser() error {
	_, err := h.Parent.DoRequest(DeleteUserPath, nil, "")
	return err
}

func (h *HubClient) CreateApp() error {
	_, err := h.Parent.DoRequest(AppCreationPath, utils.SingleString{h.App}, "")
	return err
}

func (h *HubClient) FindApps(searchTerm string) ([]UserAndApp, error) {
	result, err := h.Parent.DoRequest(SearchAppsPath, utils.SingleString{searchTerm}, "")
	if err != nil {
		return nil, err
	}

	apps, err := unpackResponse[[]UserAndApp](result)
	if err != nil {
		return nil, err
	}

	return *apps, nil
}

func (h *HubClient) GetApps() ([]string, error) {
	result, err := h.Parent.DoRequest(AppGetListPath, nil, "")
	if err != nil {
		return nil, err
	}

	apps, err := unpackResponse[[]string](result)
	if err != nil {
		return nil, err
	}

	return *apps, nil
}

func (h *HubClient) UploadTag() error {
	tapUpload := &TagUpload{
		App:     h.App,
		Tag:     h.Tag,
		Content: h.UploadContent,
	}
	_, err := h.Parent.DoRequest(TagUploadPath, tapUpload, "")
	return err
}

func (h *HubClient) DownloadTag() (string, error) {
	tagInfo := &TagInfo{
		User: h.Parent.User,
		App:  h.App,
		Tag:  h.Tag,
	}

	result, err := h.Parent.DoRequest(DownloadPath, tagInfo, "")
	if err != nil {
		return "", err
	}

	downloadedContent, ok := result.([]byte)
	if !ok {
		return "", fmt.Errorf("Failed to assert result to []byte")
	}

	return string(downloadedContent), nil
}

func (h *HubClient) GetTags() ([]string, error) {
	usernameAndApp := &UserAndApp{
		User: h.Parent.User,
		App:  h.App,
	}

	result, err := h.Parent.DoRequest(GetTagsPath, usernameAndApp, "")
	if err != nil {
		return nil, err
	}

	tags, err := unpackResponse[[]string](result)
	if err != nil {
		return nil, err
	}

	return *tags, nil
}

func unpackResponse[T any](object interface{}) (*T, error) {
	respBody, ok := object.([]byte)
	if !ok {
		return nil, fmt.Errorf("Failed to assert result to []byte")
	}

	var result T
	err := json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal response body: %v", err)
	}
	return &result, nil
}

func (h *HubClient) DeleteTag() error {
	tagInfo := &AppAndTag{
		App: h.App,
		Tag: h.Tag,
	}
	_, err := h.Parent.DoRequest(TagDeletePath, tagInfo, "")
	return err
}

func (h *HubClient) DeleteApp() error {
	_, err := h.Parent.DoRequest(AppDeletePath, utils.SingleString{h.App}, "")
	return err
}

func (h *HubClient) ChangePassword() error {
	form := utils.ChangePasswordForm{
		OldPassword: h.Parent.Password,
		NewPassword: h.Parent.NewPassword,
	}

	_, err := h.Parent.DoRequest(ChangePasswordPath, form, "")
	return err
}

func (h *HubClient) WipeData() error {
	_, err := h.Parent.DoRequest(WipeDataPath, nil, "")
	return err
}

func (h *HubClient) Logout() error {
	_, err := h.Parent.DoRequest(LogoutPath, nil, "")
	return err
}

func (h *HubClient) CheckAuth() error {
	_, err := h.Parent.DoRequest(AuthCheckPath, nil, "")
	return err
}
