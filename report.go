package revoltgo

/* This is not done || May be removed in the future */

type ReportContentMessage struct {
	ID           string `json:"id"`
	ReportReason string `json:"report_reason"`
}

type ReportContentUser struct {
	ID           string `json:"id"`
	ReportReason string `json:"report_reason"`
	MessageID    string `json:"message_id"`
}

type ReportContentServer struct {
	ID           string `json:"id"`
	ReportReason string `json:"report_reason"`
}

type ReportStatus struct {
}

type Report struct {
	ID                string `json:"id"`
	AuthorID          string `json:"author_id"`
	Content           string `json:"content"`
	AdditionalContext string `json:"additional_context"`
	Status            string `json:"status"`
	Notes             string `json:"notes"`
}
