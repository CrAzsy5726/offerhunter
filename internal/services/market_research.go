package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"offerhunter/internal/config"
	"offerhunter/internal/llm"
	"offerhunter/internal/models"
	"offerhunter/internal/prompts"
)

type MarketResearchService struct{}

func (s *MarketResearchService) Research(ctx context.Context, jobInfo, jobPreference string) (*models.MarketInfo, error) {
	searchQuery := fmt.Sprintf("%s %s 薪资 技能要求 招聘 应届生", jobInfo, jobPreference)
	searchResults := s.searchDDG(searchQuery)

	if jobPreference == "" {
		jobPreference = "不限"
	}
	if searchResults == "" {
		searchResults = "（无实时搜索结果）"
	}

	prompt := prompts.MarketResearchPrompt
	prompt = strings.ReplaceAll(prompt, "{job_info}", jobInfo)
	prompt = strings.ReplaceAll(prompt, "{job_preference}", jobPreference)
	prompt = strings.ReplaceAll(prompt, "{search_results}", searchResults)

	info := &models.MarketInfo{}
	err := callJSON(
		ctx, llm.DeepSeek(), config.DeepSeekModel, 0.3,
		prompt, prompts.MarketResearchInstruction, info,
	)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (s *MarketResearchService) searchDDG(query string) string {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://html.duckduckgo.com/html/", nil)
	if err != nil {
		return ""
	}
	q := req.URL.Query()
	q.Set("q", query)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", "OfferHunter/1.0 (job-search-tool)")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return extractSnippets(string(body))
}

var snippetRe = regexp.MustCompile(`class="result__snippet"[^>]*>(.*?)</a>`)

func extractSnippets(html string) string {
	snippets := snippetRe.FindAllStringSubmatch(html, -1)

	var lines []string
	for _, s := range snippets {
		text := stripHTMLTags(s[1])
		if len(text) > 20 {
			lines = append(lines, fmt.Sprintf("%d. %s", len(lines)+1, text))
			if len(lines) >= 5 {
				break
			}
		}
	}
	return strings.Join(lines, "\n")
}

var tagRe = regexp.MustCompile(`<[^>]*>`)

func stripHTMLTags(s string) string {
	return strings.TrimSpace(tagRe.ReplaceAllString(s, ""))
}
