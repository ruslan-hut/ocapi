package entity

type BatchResult struct {
	BatchUid   string `json:"batch_uid"`
	Success    bool   `json:"status"`
	Message    string `json:"message"`
	Products   int    `json:"products"`
	Categories int    `json:"categories"`
}
