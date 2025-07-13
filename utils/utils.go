package utils

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/ocelot-cloud/task-runner"
	"golang.org/x/crypto/bcrypt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var Logger = ProvideLogger(os.Getenv("LOG_LEVEL"))

func SendJsonResponse(w http.ResponseWriter, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		Logger.Error("unmarshalling failed: %v", err)
		http.Error(w, "failed to prepare response data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonData)
	if err != nil {
		Logger.Error("writing response failed: %v", err)
		return
	}
}

type ComponentClient struct {
	Cookie            *http.Cookie
	SetCookieHeader   bool
	RootUrl           string
	Origin            string
	VerifyCertificate bool
}

func (c *ComponentClient) DoRequest(path string, payload interface{}) ([]byte, error) {
	resp, err := c.DoRequestWithFullResponse(path, payload)
	if err != nil {
		return nil, err
	}

	respBody, err := assertOkStatusAndExtractBody(resp)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (c *ComponentClient) DoRequestWithFullResponse(path string, payload interface{}) (*http.Response, error) {
	url := c.RootUrl + path

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}
	payloadReader := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest("POST", url, payloadReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	SetCookieHeaders(req, c)
	if c.Origin != "" {
		req.Header.Set("Origin", c.Origin)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !c.VerifyCertificate}, // #nosec G402 (CWE-295): TLS InsecureSkipVerify may be true; tolerated by design
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	respBody, err := assertOkStatusAndExtractBody(resp)
	if err != nil {
		return nil, err
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
	defer Close(resp.Body)

	var bodyBuffer bytes.Buffer
	teeReader := io.TeeReader(resp.Body, &bodyBuffer)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusFound {
		respBody, err := io.ReadAll(teeReader)
		if err != nil {
			return nil, fmt.Errorf("expected status code %d, but got %d. Also failed to read response body: %v", resp.StatusCode, resp.StatusCode, err)
		}
		errorMessage := GetErrMsg(resp.StatusCode, string(respBody))
		trimmedStr := strings.TrimSuffix(errorMessage, "\n")
		return nil, errors.New(trimmedStr)
	}

	respBody, err := io.ReadAll(teeReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
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
	return fmt.Sprintf("expected status code 200, but got %d.%s", actualStatusCode, msg)
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
		Expires:  GetTimeInSevenDays(),
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
	}, nil
}

func GetTimeInSevenDays() time.Time {
	return time.Now().UTC().AddDate(0, 0, 7)
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
			file, err := os.Open(path) // #nosec G304 (CWE-22): Potential file inclusion via variable; but required by design
			if err != nil {
				return err
			}
			defer Close(file)

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		Close(zipWriter)
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
		return nil, fmt.Errorf("failed to assert result to []byte")
	}

	var result T
	err := json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}
	return &result, nil
}

var (
	maxNumberOfFilesAllowed            = 100
	maximumAllowedUnpackedBytesFromZip = int64(10 * 1024 * 1024) // 10 MB
)

// UnzipToTempDir unzips the given zip bytes to a temporary directory and returns the path to the directory.
func UnzipToTempDir(zipBytes []byte) (string, error) {
	tempDir, err := createTempDir()
	if err != nil {
		return "", err
	}
	return tempDir, unzipToDir(zipBytes, tempDir, maxNumberOfFilesAllowed, maximumAllowedUnpackedBytesFromZip)
}

func createTempDir() (string, error) {
	tempDir, err := os.MkdirTemp("", "temp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %v", err)
	}
	return tempDir, nil
}

func unzipToDir(zipBytes []byte, dest string, maxNumberOfFilesAllowed int, maxUnpackedBytesFromZip int64) error {
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return fmt.Errorf("failed to read zip file: %v", err)
	}

	var totalUnpacked int64
	if len(zipReader.File) > maxNumberOfFilesAllowed {
		return fmt.Errorf("too many files in zip: %d, max allowed: %d", len(zipReader.File), maxNumberOfFilesAllowed)
	}

	for _, file := range zipReader.File {
		if strings.Contains(file.Name, "..") {
			return fmt.Errorf("invalid file path in zip: %s", file.Name)
		}

		if err := extractFile(file, dest, &totalUnpacked, maxUnpackedBytesFromZip); err != nil {
			return err
		}
	}
	return nil
}

func extractFile(file *zip.File, dest string, totalUnpacked *int64, limit int64) error {
	fpath := filepath.Join(dest, file.Name) // #nosec G305 (CWE-22): File traversal when extracting zip/tar archive; safe due to sanitized internal file paths

	if file.FileInfo().IsDir() {
		return os.MkdirAll(fpath, 0700)
	}

	if err := os.MkdirAll(filepath.Dir(fpath), 0700); err != nil {
		return err
	}

	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer Close(rc)

	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode()) // #nosec G304 (CWE-22): File inclusion via variable path; controlled and expected
	if err != nil {
		return err
	}
	defer Close(outFile)

	// #nosec G110 (CWE-409): DoS risk via zip bomb mitigated by max unpack limit
	_, err = io.Copy(outFile, limitedCounter{rc, totalUnpacked, limit})
	return err
}

type limitedCounter struct {
	r     io.Reader
	total *int64
	limit int64
}

func (lc limitedCounter) Read(p []byte) (int, error) {
	n, err := lc.r.Read(p)
	if n > 0 {
		if *lc.total+int64(n) > lc.limit {
			return 0, fmt.Errorf("unpacked data exceeds limit")
		}
		*lc.total += int64(n)
	}
	return n, err
}

func ExecuteShellCommand(shellCommand string) error {
	return exec.Command("/bin/sh", "-c", shellCommand).Run()
}

func Execute(commandStr string) {
	commandParts := strings.Split(commandStr, " ")
	command := exec.Command(commandParts[0], commandParts[1:]...) // #nosec G204 (CWE-78): Subprocess launched with a potential tainted input or cmd arguments; but required by design
	err := command.Run()
	if err != nil {
		fmt.Printf("Error executing docker command: %v\n", err)
		os.Exit(1)
	}
}

func FindDir(dirName string) string {
	currentDir, err := os.Getwd()
	initialDir := currentDir
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
			Logger.Fatal("folder '%s' not found in any parent directory. Initial directory was: %s", dirName, initialDir)
		}

		currentDir = parentDir
	}
}

func RunMigrations(migrationsDir, host, port string) {
	m, err := migrate.New(
		"file://"+migrationsDir,
		fmt.Sprintf("postgres://postgres@%s:%s/postgres?sslmode=disable", host, port),
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
			return nil, fmt.Errorf("failed to connect to database: %v", err)
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

type Closable interface {
	Close() error
}

func Close(r Closable) {
	if err := r.Close(); err != nil {
		Logger.Error("Failed to close: %v", err)
	}
}

func RemoveDir(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		Logger.Error("Failed to delete temp directory: %v", err)
	}
}

func AnalyzeCode(tr *taskrunner.TaskRunner, dir string) {
	tr.Log.TaskDescription("Analysing code of backend")

	buildTags, err := CollectBuildTags(dir)
	if err != nil {
		tr.Log.Error("Error collecting build tags: %s", err.Error())
		os.Exit(1)
	}
	if len(buildTags) == 0 {
		tr.Log.Info("No build tags found")
	} else {
		tr.Log.Info("Build tags found: %s", strings.Join(buildTags, ", "))
	}
	buildTagString := strings.Join(buildTags, ",")

	var buildTagsInsertionString string
	if len(buildTags) == 0 {
		buildTagsInsertionString = " "
	} else {
		buildTagsInsertionString = "-tags " + buildTagString + " "
	}

	vetCmd := fmt.Sprintf("go vet %s./...", buildTagsInsertionString)
	tr.Log.Info("\ngo vet: reports suspicious constructs, such as Printf calls whose arguments do not align with the format string")
	tr.ExecuteInDir(dir, vetCmd)

	staticCheckCmd := fmt.Sprintf("staticcheck %s ./...", buildTagsInsertionString)
	tr.Log.Info("\nstaticcheck: finds bugs, performance issues, and other problems in Go code")
	tr.ExecuteInDir(dir, staticCheckCmd)

	golangCiLintCmd := fmt.Sprintf("golangci-lint run --build-tags %s ./...", buildTagString)
	tr.Log.Info("\ngolangci-lint: runs multiple linters to enforce style and catch potential bugs")
	tr.ExecuteInDir(dir, golangCiLintCmd)

	goSecCmd := fmt.Sprintf("gosec %s ./...", buildTagsInsertionString)
	tr.Log.Info("\ngosec: checks for common security issues in Go code")
	tr.ExecuteInDir(dir, goSecCmd)

	errCheckCmd := fmt.Sprintf("errcheck %s ./...", buildTagsInsertionString)
	tr.Log.Info("\nerrcheck: reports unchecked errors in Go code")
	tr.ExecuteInDir(dir, errCheckCmd)

	tr.Log.Info("\nineffassign: detects assignments to variables that are never used")
	tr.ExecuteInDir(dir, "ineffassign ./...")

	tr.Log.Info("\nunparam: reports unused function parameters")
	tr.ExecuteInDir(dir, "unparam ./...")
}

func CollectBuildTags(dir string) ([]string, error) {
	var tags []string
	tagSet := make(map[string]struct{})
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		f, err := os.Open(path) // #nosec G304 (CWE-22): Potential file inclusion via variable; but required by design
		if err != nil {
			return err
		}
		defer Close(f)
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			extractedTags := extractTagsFromLine(line)
			for _, tag := range extractedTags {
				tagSet[tag] = struct{}{}
			}
		}
		return scanner.Err()
	})
	if err != nil {
		return nil, err
	}
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags, nil
}

func extractTagsFromLine(line string) []string {
	var tags []string
	if strings.HasPrefix(line, "//go:build") {
		fields := strings.Fields(line)
		if len(fields) > 1 {
			tags = fields[1:]
		}
	}
	if strings.HasPrefix(line, "// +build") {
		fields := strings.Fields(line)
		if len(fields) > 2 {
			tags = fields[2:]
		}
	}

	var filteredTags []string
	for _, tag := range tags {
		if strings.HasPrefix(tag, "!") {
			continue
		}
		filteredTags = append(filteredTags, tag)
	}

	return filteredTags
}
