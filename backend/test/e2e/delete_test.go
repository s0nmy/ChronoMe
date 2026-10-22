package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeleteRequiresOwnedResource(t *testing.T) {
	for _, resource := range []string{"entries", "projects", "tags"} {
		t.Run(resource, func(t *testing.T) {
			fx := newFixture(t)
			fx.createSupabaseUser("delete-owner@example.com")
			ownerToken := fx.token
			var id uuid.UUID
			switch resource {
			case "entries":
				project := fx.createProject("Entry project", "#123456")
				id = fx.createEntry(project.ID, "Owned entry", time.Now().UTC(), nil).ID
			case "projects":
				id = fx.createProject("Owned project", "#123456").ID
			case "tags":
				id = fx.createTag("Owned tag").ID
			}
			path := "/api/" + resource + "/"

			status, body := fx.do(http.MethodDelete, path+uuid.NewString(), nil)
			require.Equal(t, http.StatusNotFound, status)
			require.JSONEq(t, `{"error":"resource not found"}`, string(body))

			status, _ = fx.do(http.MethodDelete, path+"invalid-id", nil)
			require.Equal(t, http.StatusBadRequest, status)

			fx.createSupabaseUser("delete-other@example.com")
			status, body = fx.do(http.MethodDelete, path+id.String(), nil)
			require.Equal(t, http.StatusNotFound, status)
			require.JSONEq(t, `{"error":"resource not found"}`, string(body))

			// 他ユーザーからの削除試行後も、所有者の記録が残っている。
			fx.token = ownerToken
			switch resource {
			case "entries":
				items := fx.listEntries()
				require.Len(t, items, 1)
				require.Equal(t, id, items[0].ID)
			case "projects":
				items := fx.listProjects()
				require.Len(t, items, 1)
				require.Equal(t, id, items[0].ID)
			case "tags":
				items := fx.listTags()
				require.Len(t, items, 1)
				require.Equal(t, id, items[0].ID)
			}

			status, body = fx.do(http.MethodDelete, path+id.String(), nil)
			require.Equal(t, http.StatusNoContent, status)
			require.Empty(t, body)

			// 削除済みのIDを再度指定した場合も、成功扱いにしない。
			status, body = fx.do(http.MethodDelete, path+id.String(), nil)
			require.Equal(t, http.StatusNotFound, status)
			require.JSONEq(t, `{"error":"resource not found"}`, string(body))
		})
	}
}
