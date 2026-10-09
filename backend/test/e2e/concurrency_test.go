package e2e

import (
	"chronome/internal/domain/entity"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestVersionedUpdates(t *testing.T) {
	fx := newFixture(t)
	fx.createSupabaseUser("concurrency@example.com")
	project := fx.createProject("original", "#111111")
	tag := fx.createTag("original")
	entry := fx.createEntry(project.ID, "original", time.Now().Add(-time.Hour), nil)
	for _, tc := range []struct {
		path, field string
		version     int64
	}{
		{"/api/entries/" + entry.ID.String(), "title", entry.Version},
		{"/api/projects/" + project.ID.String(), "name", project.Version},
		{"/api/tags/" + tag.ID.String(), "name", tag.Version},
	} {
		t.Run(tc.path, func(t *testing.T) {
			require.EqualValues(t, 1, tc.version)
			var updated map[string]any
			status := fx.doJSON(http.MethodPatch, tc.path, map[string]any{tc.field: "first", "version": tc.version}, &updated)
			require.Equal(t, http.StatusOK, status)
			require.EqualValues(t, tc.version+1, updated["version"])
			status, body := fx.do(http.MethodPatch, tc.path, map[string]any{tc.field: "stale", "version": tc.version})
			require.Equal(t, http.StatusConflict, status)
			require.Contains(t, string(body), "reload")
			// Version omission remains compatible; fields omitted by the patch survive.
			patch := map[string]any{"color": "#222222"}
			if tc.field == "title" {
				patch = map[string]any{"notes": "new notes"}
			}
			status = fx.doJSON(http.MethodPatch, tc.path, patch, &updated)
			if tc.field == "title" {
				require.Equal(t, "first", updated["title"])
			} else {
				require.Equal(t, "first", updated["name"])
			}
			require.Equal(t, http.StatusOK, status)
			require.EqualValues(t, tc.version+2, updated["version"])
			status, _ = fx.do(http.MethodPatch, tc.path, map[string]any{"version": 0})
			require.Equal(t, http.StatusBadRequest, status)
		})
	}
	var current entity.Entry
	end := time.Now().UTC().Truncate(time.Second)
	status := fx.doJSON(http.MethodPatch, "/api/entries/"+entry.ID.String(), map[string]any{"version": 3, "ended_at": end.Format(time.RFC3339)}, &current)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "first", current.Title)
	require.NotNil(t, current.EndedAt)
	require.Positive(t, current.DurationSec)
}
