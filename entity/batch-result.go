package entity

type BatchResult struct {
	BatchUid     string `json:"batch_uid"`
	Success      bool   `json:"status"`
	Message      string `json:"message"`
	Products     int    `json:"products"`
	Categories   int    `json:"categories"`
	DeletedFiles int    `json:"deleted_files"`
}

func NewBatchResult(batchUid string, err error) *BatchResult {
	if err != nil {
		return &BatchResult{
			BatchUid: batchUid,
			Success:  false,
			Message:  err.Error(),
		}
	} else {
		return &BatchResult{
			BatchUid: batchUid,
			Success:  true,
			Message:  "",
		}
	}
}
