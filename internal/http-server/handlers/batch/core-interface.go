package batch

import "ocapi/entity"

type Core interface {
	FinishBatch(batchUid string) (*entity.BatchResult, error)
}
