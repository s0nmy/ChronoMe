package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"chronome/internal/domain/entity"
)

func TestEntryTimeConsistency(t *testing.T) {
	fx := newFixture(t)
	fx.createSupabaseUser("entry-time@example.com")
	start := time.Date(2026, 10, 6, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	var entry entity.Entry
	status := fx.doJSON(http.MethodPost, "/api/entries/", map[string]string{
		"title": "Original", "started_at": start.Format(time.RFC3339), "ended_at": end.Format(time.RFC3339),
	}, &entry)
	require.Equal(t, http.StatusCreated, status)
	require.EqualValues(t, 3600, entry.DurationSec)

	for _, tc := range []struct {
		name    string
		method  string
		path    string
		payload map[string]string
	}{
		{"create equal", http.MethodPost, "/api/entries/", map[string]string{"started_at": start.Format(time.RFC3339), "ended_at": start.Format(time.RFC3339)}},
		{"create reversed", http.MethodPost, "/api/entries/", map[string]string{"started_at": end.Format(time.RFC3339), "ended_at": start.Format(time.RFC3339)}},
		{"start equals end", http.MethodPatch, "/api/entries/" + entry.ID.String(), map[string]string{"started_at": end.Format(time.RFC3339)}},
		{"start after end", http.MethodPatch, "/api/entries/" + entry.ID.String(), map[string]string{"started_at": start.Add(2 * time.Hour).Format(time.RFC3339)}},
		{"end equals start", http.MethodPatch, "/api/entries/" + entry.ID.String(), map[string]string{"ended_at": start.Format(time.RFC3339)}},
		{"end before start", http.MethodPatch, "/api/entries/" + entry.ID.String(), map[string]string{"ended_at": start.Add(-time.Hour).Format(time.RFC3339)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.payload["title"] = "Rejected"
			status, body := fx.do(tc.method, tc.path, tc.payload)
			require.Equal(t, http.StatusBadRequest, status)
			require.Contains(t, string(body), "ended_at must be after started_at")
			entries := fx.listEntries()
			require.Len(t, entries, 1)
			require.Equal(t, entry.ID, entries[0].ID)
			require.Equal(t, "Original", entries[0].Title)
			require.True(t, start.Equal(entries[0].StartedAt))
			require.NotNil(t, entries[0].EndedAt)
			require.True(t, end.Equal(*entries[0].EndedAt))
			require.EqualValues(t, 3600, entries[0].DurationSec)
		})
	}

	// 開始だけを変更したときも、保存済みの終了時刻との組み合わせで再計算する。
	var updated entity.Entry
	status = fx.doJSON(http.MethodPatch, "/api/entries/"+entry.ID.String(), map[string]string{
		"started_at": start.Add(30 * time.Minute).Format(time.RFC3339),
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.EqualValues(t, 1800, updated.DurationSec)

	// 両方を同時に動かす場合は、変更後の時刻同士を検証する。
	status = fx.doJSON(http.MethodPatch, "/api/entries/"+entry.ID.String(), map[string]string{
		"started_at": start.Add(2 * time.Hour).Format(time.RFC3339),
		"ended_at":   start.Add(4 * time.Hour).Format(time.RFC3339),
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.EqualValues(t, 7200, updated.DurationSec)
	entries := fx.listEntries()
	require.Len(t, entries, 1)
	require.True(t, start.Add(2*time.Hour).Equal(entries[0].StartedAt))
	require.EqualValues(t, 7200, entries[0].DurationSec)
}
