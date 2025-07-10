package store

import (
	"fmt"
	"github.com/ocelot-cloud/shared/utils"
)

var (
	ApiPrefix    = "/api"
	WipeDataPath = ApiPrefix + "/wipe-data"

	userPath            = ApiPrefix + "/account"
	RegistrationPath    = userPath + "/registration"
	EmailValidationPath = userPath + "/validate"
	LoginPath           = userPath + "/login"
	LogoutPath          = userPath + "/logout"
	AuthCheckPath       = userPath + "/auth-check"
	DeleteUserPath      = userPath + "/delete"
	ChangePasswordPath  = userPath + "/change-password"

	VersionPath       = ApiPrefix + "/versions"
	VersionUploadPath = VersionPath + "/upload"
	VersionDeletePath = VersionPath + "/delete"
	GetVersionsPath   = VersionPath + "/list"
	DownloadPath      = VersionPath + "/download"

	AppPath         = ApiPrefix + "/apps"
	AppCreationPath = AppPath + "/create"
	AppGetListPath  = AppPath + "/get-list"
	AppDeletePath   = AppPath + "/delete"
	SearchAppsPath  = AppPath + "/search"
)

func (h *AppStoreClient) RegisterAndValidateUser() error {
	err := h.RegisterUser()
	if err != nil {
		return err
	}
	return h.ValidateCode()
}

func (h *AppStoreClient) RegisterUser() error {
	form := getRegistrationForm(h)
	_, err := h.Parent.DoRequest(RegistrationPath, form)
	return err
}

func getRegistrationForm(hub *AppStoreClient) *RegistrationForm {
	return &RegistrationForm{
		User:     hub.Parent.User,
		Password: hub.Parent.Password,
		Email:    hub.Email,
	}
}

func (h *AppStoreClient) ValidateCode() error {
	_, err := h.Parent.DoRequest(EmailValidationPath+"?code="+h.ValidationCode, nil)
	return err
}

func (h *AppStoreClient) Login() error {
	creds := LoginCredentials{
		User:     h.Parent.User,
		Password: h.Parent.Password,
	}

	resp, err := h.Parent.DoRequestWithFullResponse(LoginPath, creds)
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

func (h *AppStoreClient) DeleteUser() error {
	_, err := h.Parent.DoRequest(DeleteUserPath, nil)
	return err
}

func (h *AppStoreClient) CreateApp() error {
	_, err := h.Parent.DoRequest(AppCreationPath, AppNameString{Value: h.App})
	if err != nil {
		return err
	}
	apps, err := h.ListOwnApps()
	if err != nil {
		return err
	}
	for _, app := range apps {
		if app.Name == h.App {
			h.AppId = app.Id
			return nil
		}
	}
	return fmt.Errorf("app not found on server")
}

func (h *AppStoreClient) SearchForApps(searchTerm string) ([]AppWithLatestVersion, error) {
	appSearchRequest := AppSearchRequest{
		SearchTerm:         searchTerm,
		ShowUnofficialApps: h.ShowUnofficialApps,
	}
	result, err := h.Parent.DoRequest(SearchAppsPath, appSearchRequest)
	if err != nil {
		return nil, err
	}

	apps, err := utils.UnpackResponse[[]AppWithLatestVersion](result)
	if err != nil {
		return nil, err
	}

	return *apps, nil
}

func (h *AppStoreClient) ListOwnApps() ([]App, error) {
	result, err := h.Parent.DoRequest(AppGetListPath, nil)
	if err != nil {
		return nil, err
	}

	apps, err := utils.UnpackResponse[[]App](result)
	if err != nil {
		return nil, err
	}

	return *apps, nil
}

func (h *AppStoreClient) UploadVersion() error {
	tapUpload := &VersionUpload{
		AppId:   h.AppId,
		Version: h.Version,
		Content: h.UploadContent,
	}
	_, err := h.Parent.DoRequest(VersionUploadPath, tapUpload)
	if err != nil {
		return err
	}

	versions, err := h.GetVersions()
	if err != nil {
		return err
	}
	for _, version := range versions {
		if version.Name == h.Version {
			h.VersionId = version.Id
			return nil
		}
	}
	return fmt.Errorf("version not found on server")
}

func (h *AppStoreClient) DownloadVersion() (*FullVersionInfo, error) {
	result, err := h.Parent.DoRequest(DownloadPath, NumberString{Value: h.VersionId})
	if err != nil {
		return nil, err
	}

	fullVersionInfo, err := utils.UnpackResponse[FullVersionInfo](result)
	if err != nil {
		return nil, err
	}

	return fullVersionInfo, nil
}

func (h *AppStoreClient) GetVersions() ([]Version, error) {
	result, err := h.Parent.DoRequest(GetVersionsPath, NumberString{Value: h.AppId})
	if err != nil {
		return nil, err
	}

	versions, err := utils.UnpackResponse[[]Version](result)
	if err != nil {
		return nil, err
	}

	return *versions, nil
}

func (h *AppStoreClient) DeleteVersion() error {
	_, err := h.Parent.DoRequest(VersionDeletePath, NumberString{Value: h.VersionId})
	return err
}

func (h *AppStoreClient) DeleteApp() error {
	_, err := h.Parent.DoRequest(AppDeletePath, NumberString{Value: h.AppId})
	return err
}

func (h *AppStoreClient) ChangePassword() error {
	form := ChangePasswordForm{
		OldPassword: h.Parent.Password,
		NewPassword: h.Parent.NewPassword,
	}

	_, err := h.Parent.DoRequest(ChangePasswordPath, form)
	return err
}

func (h *AppStoreClient) WipeData() {
	_, err := h.Parent.DoRequest(WipeDataPath, nil)
	if err != nil {
		panic("failed to wipe data: " + err.Error())
	}
}

func (h *AppStoreClient) Logout() error {
	_, err := h.Parent.DoRequest(LogoutPath, nil)
	return err
}

func (h *AppStoreClient) CheckAuth() error {
	_, err := h.Parent.DoRequest(AuthCheckPath, nil)
	return err
}

func (h *AppStoreClient) SetVersionId(versionId string) {
	h.VersionId = versionId
}

func (h *AppStoreClient) SetSearchForUnofficialApps(search bool) {
	h.ShowUnofficialApps = search
}

func (h *AppStoreClient) SetAppId(appId string) {
	h.AppId = appId
}
