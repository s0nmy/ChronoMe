# テスト戦略書

## 概要

ChronoMe のテストは「最小の労力で主要フローを守る」ことを目的とし、ユニットテストと軽量な統合テストを中心に構成します。ブラウザ操作を伴う大規模な E2E やコンテナ起動を前提とした仕組みは現段階では導入しません。

---

## テスト方針

1. **ユニットテスト優先**: ドメイン（Entity）と Usecase を対象に、インメモリのスタブやモックでロジックを検証する。  
2. **必要最低限の統合テスト**: SQLite（インメモリ or ファイル）でリポジトリ層の基本的な CRUD を確認する。将来的な移行時も同じテストが動くように抽象化する。  
3. **マンパワーで補完する領域を明確化**: UI の細かな振る舞いは手動確認で対応し、将来 Playwright 等を導入できる余地を残す。  
4. **テストデータは小さく保つ**: 再利用しやすい fixture や helper を用意し、数ケースで挙動が判断できる構成にする。

この方針に沿って、継続的な拡張が必要になったタイミングで段階的にテストを追加します。

---

## バックエンドテスト

### ユニットテスト（Usecase / Entity）

- 依存をインターフェース化し、テストではインメモリ実装やシンプルなモックを注入する。  
- 現在時刻に依存する処理は Usecase/provider の `Clock` インターフェース（`Now() time.Time`）を注入して固定値に差し替える。

```go
type StubEntryRepository struct {
    entries []*entity.Entry
}

func (s *StubEntryRepository) Create(ctx context.Context, e *entity.Entry) error {
    s.entries = append(s.entries, e)
    return nil
}

func TestStartWork(t *testing.T) {
    repo := &StubEntryRepository{}
    clock := clock.FixedClock{Fixed: time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)}
    uc := usecase.NewEntryUsecase(repo, clock)

    entry, err := uc.StartWork(ctx, userID, usecase.StartWorkRequest{Title: "実装"})
    require.NoError(t, err)
    assert.Equal(t, clock.Fixed, entry.StartedAt)
    assert.Len(t, repo.entries, 1)
}
```

### 軽量統合テスト（SQLite）

- インメモリ SQLite（`file::memory:?cache=shared`）や専用のファイル DB を用意し、マイグレーションを適用してから実行する。  
- Repository の基本的な CRUD と制約（例: `ratio` のバリデーション）を検証する。

```go
func TestEntryRepository_SQLite(t *testing.T) {
    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    require.NoError(t, err)
    require.NoError(t, db.AutoMigrate(&models.Entry{}))

    repo := repository.NewGormEntryRepository(db)
    clock := clock.FixedClock{Fixed: time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)}

    entry := entity.NewEntry(userID, "実装", clock.Now())
    require.NoError(t, repo.Create(context.Background(), entry))

    got, err := repo.FindByID(context.Background(), entry.ID)
    require.NoError(t, err)
    assert.Equal(t, entry.ID, got.ID)
}
```

---

## フロントエンドテスト

### コンポーネントテスト

- React Testing Library を使用し、ユーザーイベントに対する表示の変化を確認する。  
- API レイヤーは MSW などの軽量モックを用いてレスポンスを再現する。

```tsx
test('作業開始ボタン押下でカウントが開始する', () => {
  render(<Timer />);
  fireEvent.click(screen.getByRole('button', { name: '作業開始' }));
  expect(screen.getByText('作業中')).toBeInTheDocument();
});
```

### フック / API クライアントテスト

- TanStack Query のカスタムフックは `msw` を用いて API レスポンスをモックし、副作用が最小限であることを確認する。  
- 認証やセッションまわりは手動検証で補完する。

---

## 手動確認チェックリスト

1. 新規登録 → ログイン → プロフィール確認  
2. 作業開始・終了の一連のフロー  
3. レポート画面で日次/週次/月次の切り替え  
4. セッションタイムアウト後の再ログイン（開発中はブラウザの Cookie を削除して確認）

---

## テスト実行コマンド

```bash
# バックエンドユニットテスト
cd backend
go test ./internal/... -v

# （任意）PostgreSQL を用いた統合テスト
TEST_DATABASE_URL=postgres://chronome_test:chronome_test@localhost:5433/chronome_test?sslmode=disable go test ./internal/adapter/db/... -v

# フロントエンド
cd frontend
npm run test
```

---

## 将来の拡張候補

- testcontainers を使った PostgreSQL での統合テスト
- Playwright / Cypress によるブラウザ E2E
- GitHub Actions など CI での自動実行
- OpenAPI を用いた契約テスト

現段階では上記を実装対象外とし、必要になったタイミングでドキュメントを更新します。
