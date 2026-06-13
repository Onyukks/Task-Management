package task

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Activity is a single recorded change to a task (the activity-log feature).
type Activity struct {
	ID        uuid.UUID `json:"id"`
	Action    string    `json:"action"`
	Detail    string    `json:"detail"`
	ActorName string    `json:"actorName"`
	CreatedAt time.Time `json:"createdAt"`
}

// logActivities inserts one row per change. Best-effort: a logging failure must
// never fail the underlying task mutation, so callers ignore the error after
// logging it server-side.
func (r *Repo) logActivities(ctx context.Context, taskID, actorID uuid.UUID, entries []activityEntry) error {
	if len(entries) == 0 {
		return nil
	}
	const q = `INSERT INTO task_activities (task_id, actor_id, action, detail) VALUES ($1, $2, $3, $4)`
	batch := r.db
	for _, e := range entries {
		if _, err := batch.Exec(ctx, q, taskID, actorID, e.Action, e.Detail); err != nil {
			return err
		}
	}
	return nil
}

// ListActivity returns a task's history (newest first), joining the actor name.
func (r *Repo) ListActivity(ctx context.Context, taskID uuid.UUID) ([]Activity, error) {
	const q = `
		SELECT a.id, a.action, a.detail, u.name, a.created_at
		FROM task_activities a
		JOIN users u ON u.id = a.actor_id
		WHERE a.task_id = $1
		ORDER BY a.created_at DESC`
	rows, err := r.db.Query(ctx, q, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Activity{}
	for rows.Next() {
		var a Activity
		if err := rows.Scan(&a.ID, &a.Action, &a.Detail, &a.ActorName, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

type activityEntry struct {
	Action string
	Detail string
}

// diffActivities compares a task before and after an update and returns a
// human-readable activity entry for each field that changed.
func diffActivities(before, after *Task) []activityEntry {
	var entries []activityEntry

	if before.Title != after.Title {
		entries = append(entries, activityEntry{"title", fmt.Sprintf("Renamed to %q", after.Title)})
	}
	if before.Description != after.Description {
		entries = append(entries, activityEntry{"description", "Updated the description"})
	}
	if before.Status != after.Status {
		entries = append(entries, activityEntry{"status", fmt.Sprintf("%s → %s", before.Status, after.Status)})
	}
	if before.Priority != after.Priority {
		entries = append(entries, activityEntry{"priority", fmt.Sprintf("%s → %s", before.Priority, after.Priority)})
	}
	if !sameDue(before.DueDate, after.DueDate) {
		detail := "Cleared the due date"
		if after.DueDate != nil {
			detail = "Set due date to " + after.DueDate.Format("Jan 2, 2006")
		}
		entries = append(entries, activityEntry{"due_date", detail})
	}
	return entries
}

func sameDue(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}
