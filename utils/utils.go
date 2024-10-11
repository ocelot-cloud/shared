package utils

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ocelot-cloud/shared"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var Logger = shared.ProvideLogger(os.Getenv("LOG_LEVEL"))

const OriginHeader = "Origin"

type SingleString struct {
	Value string `json:"value"`
}

type ChangePasswordForm struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func SendJsonResponse(w http.ResponseWriter, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		Logger.Error("unmarshalling failed: %v", err)
		http.Error(w, "failed to prepare response data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

type ComponentClient struct {
	User            string
	Password        string
	NewPassword     string
	Origin          string
	Cookie          *http.Cookie
	SetOriginHeader bool
	SetCookieHeader bool
	RootUrl         string
}

func (c *ComponentClient) DoRequest(path string, payload interface{}, expectedMessage string) (interface{}, error) {
	resp, err := c.DoRequestWithFullResponse(path, payload, expectedMessage)
	if err != nil {
		return nil, err
	}

	respBody, err := assertOkStatusAndExtractBody(resp)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (c *ComponentClient) DoRequestWithFullResponse(path string, payload interface{}, expectedMessage string) (*http.Response, error) {
	url := c.RootUrl + path

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal payload: %v", err)
	}
	payloadReader := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest("POST", url, payloadReader)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	SetCookieAndOriginHeaders(req, c)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to send request: %v", err)
	}

	respBody, err := assertOkStatusAndExtractBody(resp)
	if err != nil {
		return nil, err
	}

	responseMessage, _ := strings.CutSuffix(string(respBody), "\n")
	if expectedMessage != "" && expectedMessage != responseMessage {
		return nil, fmt.Errorf("Expected response message '%s', got '%s'", expectedMessage, responseMessage)
	}

	if len(resp.Cookies()) == 1 {
		c.Cookie = resp.Cookies()[0]
	}

	// Response body can only be read once. When reading it a second time, an error occurs. So a copy is created.
	newResp := &http.Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       io.NopCloser(bytes.NewBuffer(respBody)),
	}
	return newResp, nil
}

func SetCookieAndOriginHeaders(req *http.Request, c *ComponentClient) {
	if c.SetOriginHeader {
		req.Header.Set(OriginHeader, c.Origin)
	}
	if c.SetCookieHeader && c.Cookie != nil {
		req.AddCookie(c.Cookie)
	}
}

func assertOkStatusAndExtractBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	var bodyBuffer bytes.Buffer
	teeReader := io.TeeReader(resp.Body, &bodyBuffer)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		respBody, err := io.ReadAll(teeReader)
		if err != nil {
			return nil, fmt.Errorf("Expected status code %d, but got %d. Also failed to read response body: %v", resp.StatusCode, resp.StatusCode, err)
		}
		errorMessage := GetErrMsg(resp.StatusCode, string(respBody))
		trimmedStr := strings.TrimSuffix(errorMessage, "\n")
		return nil, fmt.Errorf(trimmedStr)
	}

	respBody, err := io.ReadAll(teeReader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body: %v", err)
	}

	return respBody, nil
}

func GetErrMsg(actualStatusCode int, respBodyMsg string) string {
	var msg string
	if respBodyMsg == "" {
		msg = ""
	} else {
		msg = fmt.Sprintf(" Response body: %s", respBodyMsg)
	}
	return fmt.Sprintf("Expected status code 200, but got %d.%s", actualStatusCode, msg)
}

// GetCorsDisablingHandler This is necessary to allow cross-origin requests from the Ocelot-Cloud GUI to the Hub.
// or when testing frontend and backend components separately. The "Origin" header is managed and checked with custom logic to prevent CSRF attacks.
// TODO Rename to ApplyCorsDisable
func GetCorsDisablingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
		if r.Method == "OPTIONS" { // TODO can be removed I think
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GenerateCookie() (*http.Cookie, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		Logger.Error("Failed to generate cookie: %v", err)
		return nil, err
	}
	return &http.Cookie{
		Name:    "auth",
		Value:   hex.EncodeToString(randomBytes),
		Expires: GetTimeIn30Days(),
		Path:    "/",
	}, nil
}

func GetTimeIn30Days() time.Time {
	return time.Now().UTC().AddDate(0, 0, 30)
}

func SaltAndHash(clearText string) (string, error) {
	hashValue, err := bcrypt.GenerateFromPassword([]byte(clearText), bcrypt.DefaultCost)
	if err != nil {
		Logger.Error("Failed to hash text: %v", err)
		return "", fmt.Errorf("hashing failed")
	}
	return string(hashValue), nil
}

func Hash(clearText string) (string, error) {
	hashValue := sha256.New()
	_, err := hashValue.Write([]byte(clearText))
	if err != nil {
		Logger.Error("Failed to hash text: %v", err)
		return "", fmt.Errorf("hashing failed")
	}
	return hex.EncodeToString(hashValue.Sum(nil)), nil
}

func ZipDirectoryToBytes(dirPath string) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == dirPath {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = relPath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		zipWriter.Close()
		return nil, err
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	Logger.Info("zipped files in directory %s into a file of %v bytes", dirPath, len(buf.Bytes()))
	return buf.Bytes(), nil
}

func UnpackResponse[T any](object interface{}) (*T, error) {
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
