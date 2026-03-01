package usecase

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/marumaro-git/aws-cli-tool/internal/domain/model"
	"github.com/marumaro-git/aws-cli-tool/internal/pkg/logger"
)

type DocDBRepository interface {
	InsertEvent(ctx context.Context, event model.EventDocument) (string, error)
	GetEventsSorted(ctx context.Context) ([]model.EventDocument, error)
	UpdateUserSingle(ctx context.Context, userID string, event model.EventDocument) (*model.UserState, bool, error)
	UpdateUserBatch(ctx context.Context, userID string, events []model.EventDocument) (*model.UserState, int, error)
	GetUser(ctx context.Context, userID string) (*model.UserState, error)
	DeleteUser(ctx context.Context, userID string) error
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

func (u *DocDBUseCase) UpdateUserSingle(ctx context.Context) error {
	userID := "user_001"

	// テストデータをクリーン
	if err := u.repo.DeleteUser(ctx, userID); err != nil {
		return err
	}

	now := time.Now()
	events := []model.EventDocument{
		{
			EventType: "user_created",
			EventTime: now,
			Data:      map[string]interface{}{"name": "Alice", "email": "alice@example.com"},
		},
		{
			EventType: "user_updated",
			EventTime: now.Add(2 * time.Second),
			Data:      map[string]interface{}{"name": "Alice", "email": "alice_v2@example.com"},
		},
		{
			// 古いイベント（1秒前）→ スキップされるべき
			EventType: "user_updated",
			EventTime: now.Add(1 * time.Second),
			Data:      map[string]interface{}{"name": "Alice", "email": "alice_stale@example.com"},
		},
	}

	for i, event := range events {
		user, applied, err := u.repo.UpdateUserSingle(ctx, userID, event)
		if err != nil {
			return err
		}

		if applied {
			u.logger.Info(ctx, fmt.Sprintf("[Event %d] APPLIED: type=%s, event_time=%s -> user.email=%s, version=%d",
				i+1, event.EventType, event.EventTime.Format(time.RFC3339Nano), user.Email, user.Version))
		} else {
			u.logger.Info(ctx, fmt.Sprintf("[Event %d] SKIPPED (stale): type=%s, event_time=%s -> current.email=%s, last_updated=%s",
				i+1, event.EventType, event.EventTime.Format(time.RFC3339Nano), user.Email, user.LastUpdated.Format(time.RFC3339Nano)))
		}
	}

	// 最終状態を確認
	finalUser, err := u.repo.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	u.logger.Info(ctx, fmt.Sprintf("Final state: email=%s, version=%d (stale event was correctly rejected)", finalUser.Email, finalUser.Version))
	return nil
}

func (u *DocDBUseCase) UpdateUserBatch(ctx context.Context) error {
	userID := "user_002"

	// テストデータをクリーン
	if err := u.repo.DeleteUser(ctx, userID); err != nil {
		return err
	}

	now := time.Now()

	// === Phase 1: 初期状態を作成 ===
	u.logger.Info(ctx, "=== Phase 1: Creating initial state ===")
	initialEvent := model.EventDocument{
		EventType: "user_created",
		EventTime: now.Add(2 * time.Second),
		Data:      map[string]interface{}{"name": "Bob", "email": "bob_v2@example.com"},
	}
	_, applied, err := u.repo.UpdateUserSingle(ctx, userID, initialEvent)
	if err != nil {
		return err
	}
	if applied {
		u.logger.Info(ctx, fmt.Sprintf("Initial state created: email=%s, last_updated=%s",
			initialEvent.Data["email"], initialEvent.EventTime.Format(time.RFC3339Nano)))
	}

	// === Phase 2: 古いイベントと新しいイベントが混在するバッチを投入 ===
	u.logger.Info(ctx, "=== Phase 2: Applying batch with stale + new events ===")
	batchEvents := []model.EventDocument{
		{
			// 古いイベント（now+0s） → スキップされるべき
			EventType: "user_updated",
			EventTime: now,
			Data:      map[string]interface{}{"name": "Bob", "email": "bob_stale_1@example.com"},
		},
		{
			// 古いイベント（now+1s） → スキップされるべき
			EventType: "user_updated",
			EventTime: now.Add(1 * time.Second),
			Data:      map[string]interface{}{"name": "Bob", "email": "bob_stale_2@example.com"},
		},
		{
			// 新しいイベント（now+3s） → 適用されるべき
			EventType: "user_updated",
			EventTime: now.Add(3 * time.Second),
			Data:      map[string]interface{}{"name": "Bob", "email": "bob_v3@example.com"},
		},
		{
			// 新しいイベント（now+4s） → 適用されるべき
			EventType: "user_updated",
			EventTime: now.Add(4 * time.Second),
			Data:      map[string]interface{}{"name": "Bob", "email": "bob_v4@example.com"},
		},
	}

	u.logger.Info(ctx, "Batch events (before sort):")
	for i, e := range batchEvents {
		u.logger.Info(ctx, fmt.Sprintf("  [%d] event_time=%s, email=%s",
			i+1, e.EventTime.Format(time.RFC3339Nano), e.Data["email"]))
	}

	// event_time順にソート
	sort.Slice(batchEvents, func(i, j int) bool {
		return batchEvents[i].EventTime.Before(batchEvents[j].EventTime)
	})

	u.logger.Info(ctx, "Batch events (after sort):")
	for i, e := range batchEvents {
		u.logger.Info(ctx, fmt.Sprintf("  [%d] event_time=%s, email=%s",
			i+1, e.EventTime.Format(time.RFC3339Nano), e.Data["email"]))
	}

	user, appliedCount, err := u.repo.UpdateUserBatch(ctx, userID, batchEvents)
	if err != nil {
		return err
	}

	u.logger.Info(ctx, "=== Result ===")
	u.logger.Info(ctx, fmt.Sprintf("Batch: %d applied, %d skipped (stale) out of %d total",
		appliedCount, len(batchEvents)-appliedCount, len(batchEvents)))
	u.logger.Info(ctx, fmt.Sprintf("Final state: email=%s, version=%d", user.Email, user.Version))
	return nil
}

func validateEventOrder(events []model.EventDocument) bool {
	for i := 1; i < len(events); i++ {
		if events[i].EventTime.Before(events[i-1].EventTime) {
			return false
		}
	}
	return true
}
