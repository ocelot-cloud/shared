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
	sampleUser           = "myuser"
	sampleApp            = "myapp"
	sampleTag            = "v0.0.1"
	sampleTagFileContent = "hello"
	sampleEmail          = "testuser@example.com"
	samplePassword       = "mypassword"
	sampleOrigin         = RootUrl
	sampleForm           = &RegistrationForm{
		sampleUser,
		samplePassword,
		sampleEmail,
	}
)

type HubClient struct {
	Parent        utils.ComponentClient
	Email         string
	App           string
	Tag           string
	UploadContent []byte
}

func getRegistrationForm(hub *HubClient) *RegistrationForm {
	return &RegistrationForm{
		User:     hub.Parent.User,
		Password: hub.Parent.Password,
		Email:    hub.Email,
	}
}

func GetHub() *HubClient {
	hub := &HubClient{
		Parent: utils.ComponentClient{
			User:            sampleUser,
			Password:        samplePassword,
			Origin:          RootUrl,
			SetOriginHeader: true,
			SetCookieHeader: true,
			RootUrl:         RootUrl,
		},

		Email:         sampleEmail,
		App:           sampleApp,
		Tag:           sampleTag,
		UploadContent: []byte(sampleTagFileContent),
	}
	hub.wipeData()
	return hub
}

func (h *HubClient) registerUser() error {
	form := getRegistrationForm(h)
	_, err := h.Parent.DoRequest(RegistrationPath, form, "")
	return err
}

func (h *HubClient) login() error {
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

func (h *HubClient) deleteUser() error {
	_, err := h.Parent.DoRequest(DeleteUserPath, nil, "")
	return err
}

func (h *HubClient) createApp() error {
	_, err := h.Parent.DoRequest(AppCreationPath, utils.SingleString{h.App}, "")
	return err
}

func (h *HubClient) findApps(searchTerm string) ([]UserAndApp, error) {
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

func (h *HubClient) uploadTag() error {
	tapUpload := &TagUpload{
		App:     h.App,
		Tag:     h.Tag,
		Content: h.UploadContent,
	}
	_, err := h.Parent.DoRequest(TagUploadPath, tapUpload, "")
	return err
}

func (h *HubClient) downloadTag() (string, error) {
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

func (h *HubClient) getTags() ([]string, error) {
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

func (h *HubClient) deleteTag() error {
	tagInfo := &AppAndTag{
		App: h.App,
		Tag: h.Tag,
	}
	_, err := h.Parent.DoRequest(TagDeletePath, tagInfo, "")
	return err
}

func (h *HubClient) deleteApp() error {
	_, err := h.Parent.DoRequest(AppDeletePath, utils.SingleString{h.App}, "")
	return err
}

func (h *HubClient) changePassword() error {
	form := utils.ChangePasswordForm{
		OldPassword: h.Parent.Password,
		NewPassword: h.Parent.NewPassword,
	}

	_, err := h.Parent.DoRequest(ChangePasswordPath, form, "")
	return err
}

func (h *HubClient) wipeData() error {
	_, err := h.Parent.DoRequest(WipeDataPath, nil, "")
	return err
}

func (h *HubClient) logout() error {
	_, err := h.Parent.DoRequest(LogoutPath, nil, "")
	return err
}

func (h *HubClient) checkAuth() error {
	_, err := h.Parent.DoRequest(AuthCheckPath, nil, "")
	return err
}
