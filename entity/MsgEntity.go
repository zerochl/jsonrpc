package entity

type MsgEntity struct {
	URL        string `json:"url"`
	HTML_TEXT  string `json:"html_text"`
	COOKIE     string `json:"cookie"`
	USER_AGENT string `json:"user_agent"`
}
