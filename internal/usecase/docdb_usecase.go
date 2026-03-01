package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/marumaro-git/aws-cli-tool/internal/domain/model"
	"github.com/marumaro-git/aws-cli-tool/internal/pkg/logger"
)

type DocDBRepository interface {
	InsertEvent(ctx context.Context, event model.EventDocument) (string, error)
	GetEventsSorted(ctx context.Context) ([]model.EventDocument, error)
	Close(ctx context.Context) error
}

type DocDBUseCase struct {
	repo   DocDBRepository
	logger logger.Logger
}

func NewDocDBUseCase(repo DocDBRepository, logger logger.Logger) *DocDBUseCase {
	return &DocDBUseCase{
		repo:   repo,
		logger: logger,
	}
}

func (u *DocDBUseCase) InsertSampleEvents(ctx context.Context) error {
	events := []model.EventDocument{
		{
			EventType: "user_created",
			Data:      map[string]interface{}{"name": "Alice", "email": "alice@example.com"},
			Version:   1,
		},
		{
			EventType: "user_updated",
			Data:      map[string]interface{}{"name": "Alice", "email": "alice_new@example.com"},
			Version:   2,
		},
		{
			EventType: "user_deleted",
			Data:      map[string]interface{}{"name": "Alice"},
			Version:   3,
		},
	}

	for _, event := range events {
		id, err := u.repo.InsertEvent(ctx, event)
		if err != nil {
			return err
		}
		u.logger.Info(ctx, fmt.Sprintf("Inserted event: id=%s, type=%s", id, event.EventType))
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

func (u *DocDBUseCase) CheckEventualConsistency(ctx context.Context, maxWait time.Duration) error {
	startTime := time.Now()

	for time.Since(startTime) < maxWait {
		events, err := u.repo.GetEventsSorted(ctx)
		if err != nil {
			return err
		}

		if len(events) == 0 {
			u.logger.Info(ctx, "No events found, waiting...")
			time.Sleep(100 * time.Millisecond)
			continue
		}

		u.logger.Info(ctx, fmt.Sprintf("Found %d events, checking order...", len(events)))

		if validateEventOrder(events) {
			u.logger.Info(ctx, "Eventual consistency achieved: all events are in chronological order")
			for _, e := range events {
				u.logger.Info(ctx, fmt.Sprintf("  [%s] %s (version=%d)", e.ID, e.EventType, e.Version))
			}
			return nil
		}

		u.logger.Info(ctx, "Events not yet in order, waiting...")
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("eventual consistency not achieved within %v", maxWait)
}

func validateEventOrder(events []model.EventDocument) bool {
	for i := 1; i < len(events); i++ {
		if events[i].EventTime.Before(events[i-1].EventTime) {
			return false
		}
	}
	return true
}
