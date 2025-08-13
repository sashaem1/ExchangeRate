package internal

import (
	"context"
	"fmt"
	"time"
)

type ActionLogID string

type ActionLog struct {
	ID            ActionLogID
	ActionLogType ActionLogType
	Timestamp     time.Time
}

func NewActionLog(typeName string, timestamp time.Time) (ActionLog, error) {
	op := "internal.logDb.NewActionLog"
	logDbType, err := NewActionLogType(typeName)

	if err != nil {
		return ActionLog{}, fmt.Errorf("%s: %s", op, err)
	}

	logDb := ActionLog{
		ActionLogType: logDbType,
		Timestamp:     timestamp,
	}

	return logDb, nil
}

type ActionLogStorage interface {
	Set(ctx context.Context, ActionLog ActionLog) error
}

type ActionLogRepository struct {
	storage ActionLogStorage
}

func NewActionLogRepository(storage ActionLogStorage) *ActionLogRepository {
	return &ActionLogRepository{
		storage: storage,
	}
}

func (rr *ActionLogRepository) InsertLog(ctx context.Context, ActionLog ActionLog) error {
	op := "internal.logDb.InsertLog"

	err := rr.storage.Set(ctx, ActionLog)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	return nil
}
