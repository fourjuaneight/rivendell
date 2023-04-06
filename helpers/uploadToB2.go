package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type B2AuthResp struct {
	AbsoluteMinimumPartSize int    `json:"absoluteMinimumPartSize"`
	AccountId               string `json:"accountId"`
	Allowed                 struct {
		BucketId     string   `json:"bucketId"`
		BucketName   string   `json:"bucketName"`
		Capabilities []string `json:"capabilities"`
		NamePrefix   string   `json:"namePrefix"`
	} `json:"allowed"`
	ApiUrl              string `json:"apiUrl"`
	AuthorizationToken  string `json:"authorizationToken"`
	DownloadUrl         string `json:"downloadUrl"`
	RecommendedPartSize int    `json:"recommendedPartSize"`
	S3ApiUrl            string `json:"s3ApiUrl"`
}

type B2UpUrlResp struct {
	BucketId           string `json:"bucketId"`
	UploadUrl          string `json:"uploadUrl"`
	AuthorizationToken string `json:"authorizationToken"`
}

type B2UploadResp struct {
	FileId        string `json:"fileId"`
	FileName      string `json:"fileName"`
	AccountId     string `json:"accountId"`
	BucketId      string `json:"bucketId"`
	ContentLength int    `json:"contentLength"`
	ContentSha1   string `json:"contentSha1"`
	ContentType   string `json:"contentType"`
	FileInfo      struct {
		Author string `json:"author"`
	} `json:"fileInfo"`
	ServerSideEncryption struct {
		Algorithm string `json:"algorithm"`
		Mode      string `json:"mode"`
	} `json:"serverSideEncryption"`
}

type B2Error struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type B2AuthTokens struct {
	ApiUrl              string
	AuthorizationToken  string
	DownloadUrl         string
	RecommendedPartSize int
}

type B2UploadTokens struct {
	Endpoint    string
	AuthToken   string
	DownloadUrl string
}

// Get B2 keys from .env file.
func getKeys(key string) string {
	envPath := os.Getenv("GOPATH") + "/.env"
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatal("[getKeys]: %w", err)
	}

	APP_KEY_ID := os.Getenv("B2_APP_KEY_ID")
	APP_KEY := os.Getenv("B2_APP_KEY")
	BUCKET_ID := os.Getenv("B2_BUCKET_ID")
	BUCKET_NAME := os.Getenv("B2_BUCKET_NAME")

	keys := map[string]string{
		"APP_KEY_ID":  APP_KEY_ID,
		"APP_KEY":     APP_KEY,
		"BUCKET_ID":   BUCKET_ID,
		"BUCKET_NAME": BUCKET_NAME,
	}

	return keys[key]
}

// Authorize B2 bucket for upload.
// DOCS: https://www.backblaze.com/b2/docs/b2_authorize_account.html
func authTokens() B2AuthTokens {
	token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", getKeys("APP_KEY_ID"), getKeys("APP_KEY"))))
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.backblazeb2.com/b2api/v2/b2_authorize_account", nil)
	if err != nil {
		log.Fatal("[authTokens][http.NewRequest]: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", token))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("[authTokens][client.Do]: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var b2Error B2Error
		err := json.NewDecoder(resp.Body).Decode(&b2Error)
		if err != nil {
			log.Fatal("[authTokens][json.NewDecoder](b2Error): %w", err)
		}

		msg := b2Error.Message
		if msg == "" {
			msg = fmt.Sprintf("%d - %s", b2Error.Status, b2Error.Code)
		}
		log.Fatal("[authTokens][b2Error]: %w", msg)
	}

	var results B2AuthResp
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		log.Fatal("[authTokens][json.NewDecoder](results): %w", err)
	}

	authTokens := B2AuthTokens{
		ApiUrl:              results.ApiUrl,
		AuthorizationToken:  results.AuthorizationToken,
		DownloadUrl:         results.DownloadUrl,
		RecommendedPartSize: results.RecommendedPartSize,
	}

	return authTokens
}

// Get B2 endpoint for upload.
// DOCS: https://www.backblaze.com/b2/docs/b2_get_upload_url.html
func getUploadUrl() B2UploadTokens {
	authData := authTokens()

	payload := map[string]string{"bucketId": getKeys("BUCKET_ID")}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/b2api/v1/b2_get_upload_url", authData.ApiUrl), bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Fatal("[getUploadUrl][http.NewRequest]: %w", err)
	}
	req.Header.Set("Authorization", authData.AuthorizationToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("[getUploadUrl][client.Do]: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var b2Error B2Error
		err := json.NewDecoder(resp.Body).Decode(&b2Error)
		if err != nil {
			log.Fatal("[getUploadUrl][json.NewDecoder](b2Error): %w", err)
		}

		msg := b2Error.Message
		if msg == "" {
			msg = fmt.Sprintf("%d - %s", b2Error.Status, b2Error.Code)
		}
		log.Fatal("[getUploadUrl][b2Error]: %w", msg)
	}

	var results B2UpUrlResp
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		log.Fatal("[getUploadUrl][json.NewDecoder](results): %w", err)
	}

	uploadTokens := B2UploadTokens{
		Endpoint:    results.UploadUrl,
		AuthToken:   results.AuthorizationToken,
		DownloadUrl: authData.DownloadUrl,
	}

	return uploadTokens
}

// Upload file to B2 bucket.
// DOCS: https://www.backblaze.com/b2/docs/b2_upload_file.html
func uploadToB2(data []byte, name, fileType string) string {
	authData := getUploadUrl()

	hasher := sha1.New()
	hasher.Write(data)
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	if fileType == "" {
		fileType = "b2/x-auto"
	}

	req, err := http.NewRequest("POST", authData.Endpoint, bytes.NewReader(data))
	if err != nil {
		log.Fatal("[uploadToB2][http.NewRequest]: %w", err)
	}
	req.Header.Set("Authorization", authData.AuthToken)
	req.Header.Set("X-Bz-File-Name", name)
	req.Header.Set("Content-Type", fileType)
	req.Header.Set("Content-Length", strconv.Itoa(len(data)))
	req.Header.Set("X-Bz-Content-Sha1", hash)
	req.Header.Set("X-Bz-Info-Author", "gh-action")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("[uploadToB2][client.Do]: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var b2Error B2Error
		err := json.NewDecoder(resp.Body).Decode(&b2Error)
		if err != nil {
			log.Fatal("[uploadToB2][json.NewDecoder](b2Error): %w", err)
		}

		msg := b2Error.Message
		if msg == "" {
			msg = fmt.Sprintf("%d - %s", b2Error.Status, b2Error.Code)
		}
		log.Fatal("[uploadToB2][b2Error]: %w", msg)
	}

	var results B2UploadResp
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		log.Fatal("[uploadToB2][json.NewDecoder](results): %w", err)
	}

	log.Printf("[uploadContentB2] [uploadToB2]: Uploaded '%s'.\n", results.FileName)

	publicURL := fmt.Sprintf("%s/file/%s/%s", authData.DownloadUrl, getKeys("BUCKET_NAME"), results.FileName)

	return publicURL
}