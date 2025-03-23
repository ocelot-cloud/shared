package utils

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/ocelot-cloud/shared"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var Logger = shared.ProvideLogger(os.Getenv("LOG_LEVEL"))

type SingleString struct {
	Value string `json:"value"`
}

type SingleInteger struct {
	Value int `json:"value"`
}

// TODO to be removed
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
	User              string
	Password          string
	NewPassword       string
	Cookie            *http.Cookie
	SetCookieHeader   bool
	RootUrl           string
	Origin            string
	VerifyCertificate bool
}

func (c *ComponentClient) DoRequest(path string, payload interface{}, expectedMessage string) ([]byte, error) {
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
	SetCookieHeaders(req, c)
	if c.Origin != "" {
		req.Header.Set("Origin", c.Origin)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !c.VerifyCertificate},
		},
	}
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

func SetCookieHeaders(req *http.Request, c *ComponentClient) {
	if c.SetCookieHeader && c.Cookie != nil {
		req.AddCookie(c.Cookie)
	}
}

func assertOkStatusAndExtractBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	var bodyBuffer bytes.Buffer
	teeReader := io.TeeReader(resp.Body, &bodyBuffer)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusFound {
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

// GetCorsDisablingHandler This is necessary to allow cross-origin requests.
func GetCorsDisablingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
		if r.Method == "OPTIONS" {
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
		Name:     "auth",
		Value:    hex.EncodeToString(randomBytes),
		Expires:  GetTimeIn30Days(),
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
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

func DoesMatchSaltedHash(clearText, saltedHash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(saltedHash), []byte(clearText)) == nil
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

// TODO To be tested
// UnzipToTempDir unzips the given zip bytes to a temporary directory and returns the path to the directory.
func UnzipToTempDir(zipBytes []byte) (string, error) {
	tempDir, err := createTempDir()
	if err != nil {
		return "", err
	}
	return tempDir, unzipToDir(zipBytes, tempDir)
}

func createTempDir() (string, error) {
	tempDir, err := os.MkdirTemp("", "temp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %v", err)
	}
	return tempDir, nil
}

func unzipToDir(zipBytes []byte, dest string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return fmt.Errorf("failed to read zip file: %v", err)
	}

	for _, file := range zipReader.File {
		if strings.Contains(file.Name, "..") {
			return fmt.Errorf("invalid file path in zip: %s", file.Name)
		}

		fpath := filepath.Join(dest, file.Name)
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %v", err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to open file: %v", err)
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return fmt.Errorf("failed to copy file: %v", err)
		}

		rc.Close()
		outFile.Close()
	}
	return nil
}

func ExecuteShellCommand(shellCommand string) error {
	return exec.Command("/bin/sh", "-c", shellCommand).Run()
}

func Execute(commandStr string) {
	commandParts := strings.Split(commandStr, " ")
	command := exec.Command(commandParts[0], commandParts[1:]...)
	err := command.Run()
	if err != nil {
		fmt.Printf("Error executing docker command: %v\n", err)
		os.Exit(1)
	}
}

func FindDir(dirName string) string {
	currentDir, err := os.Getwd()
	if err != nil {
		Logger.Fatal("Failed to get current dir: %v", err)
	}

	for {
		assetsPath := filepath.Join(currentDir, dirName)
		if _, err := os.Stat(assetsPath); err == nil {
			return assetsPath
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir { // Reached root folder
			Logger.Fatal("Assets folder not found in any parent directory")
		}

		currentDir = parentDir
	}
}

func RunMigrations(migrationsDir, host string) {
	m, err := migrate.New(
		"file://"+migrationsDir,
		fmt.Sprintf("postgres://postgres@%s:5432/postgres?sslmode=disable", host),
	)
	if err != nil {
		Logger.Fatal("Migration init failed: %v", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		Logger.Fatal("Migration failed: %v", err)
	}
}

func WaitForPostgresDb(host, port string) (*sql.DB, error) {
	var err error
	var dbClient *sql.DB
	counter := 0
	attempts := 30
	for {
		counter++
		if counter >= attempts {
			return nil, fmt.Errorf("Failed to connect to database: %v", err)
		}

		dataSourceName := fmt.Sprintf("host=%s port=%s user=postgres dbname=postgres sslmode=disable", host, port)
		dbClient, err = sql.Open("postgres", dataSourceName)
		if err == nil {
			err = dbClient.Ping()
			if err == nil {
				break // DB is ready
			}
		}

		Logger.Info("Waiting for postgres database at host '%s' to be ready... (%d/%d)", host, counter, attempts)
		time.Sleep(1 * time.Second)
	}
	return dbClient, nil
}
