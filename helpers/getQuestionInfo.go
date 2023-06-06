package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type QuestionOwner struct {
	AccountID    int    `json:"account_id"`
	Reputation   int    `json:"reputation"`
	UserID       int    `json:"user_id"`
	UserType     string `json:"user_type"`
	AcceptRate   int    `json:"accept_rate"`
	ProfileImage string `json:"profile_image"`
	DisplayName  string `json:"display_name"`
	Link         string `json:"link"`
}

type QuestionItems struct {
	Tags               []string      `json:"tags"`
	Owner              QuestionOwner `json:"owner"`
	IsAnswered         bool          `json:"is_answered"`
	ViewCount          int           `json:"view_count"`
	ProtectedDate      int           `json:"protected_date"`
	AcceptedAnswerID   int           `json:"accepted_answer_id"`
	AnswerCount        int           `json:"answer_count"`
	CommunityOwnedDate int           `json:"community_owned_date"`
	Score              int           `json:"score"`
	LockedDate         int           `json:"locked_date"`
	LastActivityDate   int           `json:"last_activity_date"`
	CreationDate       int           `json:"creation_date"`
	LastEditDate       int           `json:"last_edit_date"`
	QuestionID         int           `json:"question_id"`
	ContentLicense     string        `json:"content_license"`
	Link               string        `json:"link"`
	Title              string        `json:"title"`
}

type QuestionResponse struct {
	Items          []QuestionItems `json:"items"`
	HasMore        bool            `json:"has_more"`
	QuotaMax       int             `json:"quota_max"`
	QuotaRemaining int             `json:"quota_remaining"`
}

type CleanQuestion struct {
	Title    string
	Question string
	Answer   string
	Tags     []string
}

func parseSEURL(url string) (string, string, error) {
	regex, err := regexp.Compile(`.*(askubuntu|serverfault|stackoverflow|superuser)\.com/questions/([^/]+)/([^/]+)`)
	if err != nil {
		return "", "", fmt.Errorf("[parseSEURL][regexp.Compile]: %w", err)

	}

	matches := regex.FindStringSubmatch(url)
	if len(matches) == 3 {
		return matches[1], matches[2], nil
	}

	return "", "", fmt.Errorf("[parseSEURL]: No matches found%w", nil)
}

// DOCS: https://api.stackexchange.com/docs/questions-by-ids#order=desc&sort=activity&ids=34230208&filter=default&site=stackoverflow&run=true
func GetQuestionInfo(url string) (CleanQuestion, error) {
	site, id, err := parseSEURL(url)
	if err != nil {
		return CleanQuestion{}, fmt.Errorf("[GetQuestionInfo]%w", err)
	}

	queryURL := fmt.Sprintf("https://api.stackexchange.com/2.3/questions/%s?order=desc&sort=activity&site=%s", id, site)

	resp, err := http.Get(queryURL)
	if err != nil {
		return CleanQuestion{}, fmt.Errorf("[GetQuestionInfo][http.NewRequest]: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CleanQuestion{}, fmt.Errorf("[GetQuestionInfo][io.ReadAll]: %w", err)
	}

	var questionResponse QuestionResponse
	err = json.Unmarshal(body, &questionResponse)
	if err != nil {
		return CleanQuestion{}, fmt.Errorf("[GetQuestionInfo][json.Unmarshal]: %w", err)
	}

	questionItem := questionResponse.Items[0]
	question := fmt.Sprintf("https://%s.com/q/%d", site, questionItem.QuestionID)
	answer := ""
	if questionItem.IsAnswered {
		answer = fmt.Sprintf("https://%s.com/a/%d", site, questionItem.AcceptedAnswerID)
	}

	return CleanQuestion{
		Title:    questionItem.Title,
		Question: question,
		Answer:   answer,
		Tags:     questionItem.Tags,
	}, nil
}
