# ChronoMe API 設計書

最終更新: 2025-10-16（執筆日で更新してください）

本ドキュメントは「docs/DesignDoc.md」の技術設計を基に、ChronoMe（React + TypeScript フロントエンド、Go バックエンド）の REST API を詳細化したものです。開発者および QA が共通認識を持つための仕様書として利用します。

---

## 1. API 概要
- **API スタイル**: JSON over HTTPS (RESTFul)
- **ベース URL**: `https://api.chronome.example.com/v1`
- **文字コード**: UTF-8
- **データフォーマット**: `application/json`
- **タイムゾーン管理**: DB は UTC 固定で保存。レスポンスはユーザー設定（`tz`）がない場合 UTC を返却。クライアントはローカル表示で変換する。
- **日付形式**: ISO 8601 (`YYYY-MM-DDTHH:MM:SSZ`)
- **ID**: 全て UUID v4
- **リトライ**: クライアント実装に委ねる（Idempotent な `GET`/`PUT`/`DELETE` のみに限定）
- **ステータスコード方針**:
  - `2xx`: 成功 (`200 OK`, `201 Created`, `204 No Content`)
  - `4xx`: クライアントエラー (`400`, `401`, `403`, `404`, `409`, `422`)
- `5xx`: サーバーエラー（想定外は `500`。リカバリ可能な一時的障害は `503`）

> データベース列・制約の詳細は `docs/DBDesign.md` を参照してください。

---

## 2. 認証・認可
- **方式**: サインド Cookie ベースのセッション。ログイン時にユーザーIDと有効期限を含むトークンを生成し、HMAC-SHA256 で署名した Cookie を発行する（サーバー側に状態は保持しない）。
- **セッション寿命**: 12 時間。延長処理は設けず、期限切れ後は再ログインで対応する。
- **CSRF 対策**: `SameSite=Lax` の Cookie 設定を採用し、状態変更エンドポイントでは `POST/PUT/PATCH/DELETE` のみを使用する。
- **認可**: リクエストが保持するセッションのユーザー ID と一致するデータのみ操作可能。Usecase 層で所有者チェックを行う。

---

## 3. 共通仕様

### 3.1 エラーレスポンス
```json
{
  "error": {
    "code": "ENTRY_NOT_FOUND",
    "message": "対象のタイムエントリが存在しません。",
    "details": {
      "entry_id": "..."
    }
  }
}
```
- `code`: 定数（`AUTH_INVALID_CREDENTIALS`, `PROJECT_CONFLICT_NAME`, 等）
- `message`: ローカライズ可能なユーザー向け文言（日本語/英語切り替え想定）
- `details`: 任意（フィールドエラーや補足情報をマップで返却）

### 3.2 ページネーション
- **クエリ**: `?page=1&per_page=20`
- **レスポンスヘッダ**: `X-Total-Count`, `X-Total-Pages`
- **最大 per_page**: 100

### 3.3 並行更新制御
- **方式**: `If-Match` ヘッダ + エンティティ側の `version`（`updated_at` でも可）。一致しない場合 `409 Conflict`。

### 3.4 ソート/フィルタ
- `GET /api/entries`: `?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z&project_id=...&tag_id=...&sort=-started_at`
- `GET /api/projects`: `?sort=name&include_archived=true`

---

## 4. モデル定義

### 4.1 User
> 対応テーブル: `docs/DBDesign.md` 「4.1 users」

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `id` | string(UUID) | ユーザー ID |
| `email` | string | 一意メールアドレス |
| `display_name` | string | 表示名（任意） |
| `time_zone` | string | IANA timezone（未設定時は `UTC`） |
| `created_at` | string(datetime) | 登録日時 |
| `updated_at` | string(datetime) | 更新日時 |

### 4.2 Project
> 対応テーブル: `docs/DBDesign.md` 「4.2 projects」

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `id` | string(UUID) | プロジェクト ID |
| `user_id` | string(UUID) | 所有ユーザー |
| `name` | string | プロジェクト名（ユーザー内ユニーク） |
| `color` | string | HEX カラー（`#RRGGBB`） |
| `is_archived` | boolean | アーカイブ済みか（一覧では既定非表示） |
| `created_at` | string(datetime) | 作成日時 |
| `updated_at` | string(datetime) | 更新日時 |

### 4.3 Tag
> 対応テーブル: `docs/DBDesign.md` 「4.3 tags」

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `id` | string(UUID) | タグ ID |
| `user_id` | string(UUID) | 所有者 |
| `name` | string | タグ名 |
| `color` | string | HEX カラー |
| `created_at` / `updated_at` | string(datetime) | 作成・更新日時 |

### 4.4 Entry
> 対応テーブル: `docs/DBDesign.md` 「4.4 entries」

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `id` | string(UUID) | エントリ ID |
| `user_id` | string(UUID) | 所有者 |
| `project_id` | string(UUID) | 紐付くプロジェクト |
| `title` | string | 作業タイトル |
| `started_at` | string(datetime) | 開始時刻（UTC） |
| `ended_at` | string(datetime) | 終了時刻（未終了は `null`） |
| `duration_sec` | number | 所要秒数（バックエンド算出） |
| `is_break` | boolean | 休憩扱いか |
| `ratio` | number | 並行作業割合（0.0〜1.0、小数第 2 位まで） |
| `notes` | string | 備考 |
| `tags` | Tag[] | 紐付タグ一覧 |
| `created_at` / `updated_at` | string(datetime) | 作成・更新日時 |

### 4.5 ReportSummary
> 集計は `docs/DBDesign.md` 「5. ビュー / マテリアライズドビュー（任意提案）」を参照

| フィールド | 型 | 説明 |
| --- | --- | --- |
| `range` | object | `{"from": "...", "to": "..."}` |
| `total_duration_sec` | number | 合計稼働時間 |
| `billable_duration_sec` | number | 請求対象時間（`is_break=false` のみ） |
| `projects` | array | プロジェクト別集計（`project_id`, `name`, `duration_sec`, `ratio_sum`） |
| `tags` | array | タグ別集計（`tag_id`, `name`, `duration_sec`） |

---

## 5. エンドポイント詳細

### 5.1 認証

#### POST /api/auth/signup
- **概要**: 新規ユーザー登録
- **認証**: 不要
- **リクエスト**
```json
{
  "email": "user@example.com",
  "password": "P@ssw0rd!",
  "display_name": "Miyu",
  "time_zone": "Asia/Tokyo"
}
```
- **バリデーション**
  - `password`: 8〜128 文字、英大小＋数字必須
  - `time_zone`: IANA 名称
- **レスポンス `201 Created`**
```json
{
  "user": { ...User }
}
```
- **エラー**
  - `409 Conflict`: 既存メール
  - `422 Unprocessable Entity`: バリデーション失敗

#### POST /api/auth/login
- **概要**: ログインしセッション発行
- **リクエスト**
```json
{ "email": "user@example.com", "password": "P@ssw0rd!" }
```
- **レスポンス `200 OK`**
```json
{
  "user": { ...User }
}
```
- Cookie に `chronome_session`、ヘッダ `Set-Cookie: HttpOnly; Secure; SameSite=Lax`
- **エラー**: `401 Unauthorized` (`AUTH_INVALID_CREDENTIALS`)

#### POST /api/auth/logout
- **概要**: セッション破棄
- **認証**: 必須
- **レスポンス**: `204 No Content`

#### GET /api/auth/me
- **概要**: 現在のユーザー情報取得
- **レスポンス `200 OK`**
```json
{ "user": { ...User } }
```

### 5.2 プロジェクト

#### GET /api/projects
- **概要**: 自分のプロジェクト一覧
- **クエリ**: `page`, `per_page`, `sort`, `include_archived`（デフォルト `false`）
- **レスポンス `200 OK`**
```json
{
  "projects": [
    {
      "id": "...",
      "name": "Client A",
      "color": "#FFAA00",
      "is_archived": false,
      "created_at": "2024-01-02T03:04:05Z",
      "updated_at": "2024-01-05T06:07:08Z"
    }
  ]
}
```

#### POST /api/projects
- **概要**: プロジェクト作成
- **リクエスト**
```json
{ "name": "Client A", "color": "#FFAA00" }
```
- **レスポンス `201 Created`**
```json
{ "project": { ...Project } }
```
- **エラー**
  - `409 Conflict`: 同名存在 (`PROJECT_CONFLICT_NAME`)

#### PATCH /api/projects/{project_id}
- **概要**: プロジェクト更新
- **ヘッダ**: `If-Match: "updated_at"`（推奨）
- **リクエスト**
```json
{ "name": "Client Alpha", "color": "#00AAFF", "is_archived": true }
```
- **レスポンス `200 OK`**: 更新後オブジェクト
- **エラー**: `404 Not Found`, `409 Conflict`

#### DELETE /api/projects/{project_id}
- **概要**: プロジェクト削除
- **制約**: 紐付エントリが存在する場合 `409`。アーカイブのみ行う場合は `PATCH` で `is_archived=true` を設定する。
- **レスポンス**: `204 No Content`

---

### 5.3 タグ

#### GET /api/tags
- **概要**: タグ一覧（UI オートコンプリート用）
- **レスポンス**
```json
{ "tags": [ { ...Tag } ] }
```

#### GET /api/tags
- **概要**: ユーザーのタグ一覧取得
- **レスポンス `200 OK`**
```json
{ "tags": [{ "id": "...", "name": "Deep Work", "color": "#F97316" }] }
```

#### POST /api/tags
- **概要**: タグ作成
- **リクエスト**
```json
{ "name": "Deep Work", "color": "#F97316" }
```
- `color` 省略時はサーバー側デフォルト (`DEFAULT_PROJECT_COLOR`)
- **レスポンス**: `201 Created`

#### PATCH /api/tags/{tag_id}
- **概要**: タグ名/色の更新
- **バリデーション**: `color` は `#RRGGBB`
- **レスポンス `200 OK`**: 更新後の Tag

#### DELETE /api/tags/{tag_id}
- **レスポンス `204 No Content`**
- **備考**: 今後 `entry_tags` を cascade delete 予定

---

### 5.4 タイムエントリ

#### GET /api/entries
- **概要**: 期間内エントリ検索
- **クエリ**
  - `from`, `to`: 必須
  - `project_id`, `tag_id`, `is_break`, `running_only`（未終了のみ）
  - `page`, `per_page`, `sort`
- **レスポンス `200 OK`**
```json
{
  "entries": [
    {
      "id": "...",
      "title": "DesignMeeting",
      "started_at": "2024-01-01T01:00:00Z",
      "ended_at": "2024-01-01T02:00:00Z",
      "duration_sec": 3600,
      "project": { "id": "...", "name": "Client A" },
      "tags": [
        { "id": "...", "name": "Meeting" }
      ],
      "is_break": false,
      "ratio": 1.0,
      "notes": "Discussed MVP scope"
    }
  ]
}
```

#### POST /api/entries
- **概要**: 新規エントリ開始/登録
- **ビジネスルール**
  - `ended_at` が `null` の既存エントリが存在する場合、自動で `ended_at = now`, `duration_sec` を更新し `ratio=1.0` を維持。
  - `started_at` と `ended_at` が両方指定された場合、`duration_sec` はサーバーで算出し 0 以下は `422`。
- **リクエスト**
```json
{
  "project_id": "...",
  "title": "Draft Spec",
  "started_at": "2024-01-01T03:00:00Z",
  "ended_at": null,
  "is_break": false,
  "tag_ids": ["...", "..."],
  "notes": "Initial drafting"
}
```
- **レスポンス `201 Created`**: 作成後の Entry

#### PATCH /api/entries/{entry_id}
- **概要**: エントリ更新（終了・内容変更）
- **リクエスト例**
```json
{
  "ended_at": "2024-01-01T04:10:00Z",
  "is_break": false,
  "tag_ids": ["..."],
  "ratio": 0.5,
  "notes": "Parallel with research"
}
```
- **レスポンス `200 OK`**
- **備考**: `ratio` の合計が 1.0 を超える場合は `422`。Usecase 層で同期間の他エントリと集計。

#### DELETE /api/entries/{entry_id}
- **概要**: エントリ削除
- **レスポンス**: `204 No Content`
- **備考**: レポート集計整合性のため論理削除に切替える方針も検討可能。

#### POST /api/entries/start
- **概要**: 新しい作業を開始
- **リクエストボディ**:
```json
{
  "title": "新機能開発",
  "project_id": "123e4567-e89b-12d3-a456-426614174000",
  "is_break": false,
  "notes": "ガントチャート機能の実装"
}
```
- **レスポンス `201 Created`**: 作成されたEntry
- **備考**: 
  - 未終了エントリがある場合でも新規作成を許可（並行作業対応）
  - `ratio` は初期値 1.0 で作成、後から調整可能

#### POST /api/entries/{entry_id}/stop
- **概要**: 指定した作業を終了
- **リクエストボディ**:
```json
{
  "notes": "実装完了、テスト待ち"
}
```
- **レスポンス `200 OK`**: 更新されたEntry
- **エラー**: 
  - 既に終了済み: `409 Conflict`
  - 他ユーザーのエントリ: `403 Forbidden`

#### PATCH /api/entries/batch-ratio
- **概要**: 複数エントリの時間配分（ratio）を一括更新
- **用途**: ガントチャート形式UIから並行作業の時間配分を調整
- **リクエストボディ**:
```json
{
  "updates": [
    {"entry_id": "123e4567-e89b-12d3-a456-426614174000", "ratio": 0.6},
    {"entry_id": "123e4567-e89b-12d3-a456-426614174001", "ratio": 0.4}
  ]
}
```
- **レスポンス `200 OK`**: 更新されたEntryの配列
- **バリデーション**: 
  - ratio の合計が 1.0 でない場合: `422 Unprocessable Entity`
  - 存在しないentry_id: `404 Not Found`
  - 他ユーザーのエントリ: `403 Forbidden`

---

### 5.5 レポート

#### GET /api/reports/daily
- **概要**: 指定日の集計
- **クエリ**: `date=2024-01-01`（必須）, `time_zone`（任意、なければユーザー設定）
- **レスポンス `200 OK`**
```json
{
  "summary": { ...ReportSummary },
  "entries": [ ...Entry ] // 任意、ダウンロード用
}
```

#### GET /api/reports/weekly
- **クエリ**: `week_start=2024-01-01`（省略時は当週の月曜）
- **レスポンス `200 OK`**
```json
{
  "week_start": "2024-01-01",
  "total_seconds": 14400,
  "days": [
    {"date": "2024-01-01", "total_seconds": 7200},
    {"date": "2024-01-02", "total_seconds": 3000}
  ]
}
```

#### GET /api/reports/monthly
- **クエリ**: `month=2024-01`
- **レスポンス `200 OK`**
```json
{
  "month": "2024-01",
  "total_seconds": 54000,
  "days_in_month": 31,
  "days": [{ "date": "2024-01-01", "total_seconds": 3600 }],
  "weeks": [{ "week_start": "2023-12-30", "total_seconds": 7200 }],
  "projects": [{ "project_id": "123e4567-e89b-12d3-a456-426614174000", "total_seconds": 2400 }]
}
```

#### GET /api/reports/export
- **概要**: CSV / JSON エクスポート
- **クエリ**: `format=csv`（デフォルト `json`）、`from`, `to`
- **レスポンス**: `text/csv` もしくは JSON。

---

### 5.6 健康診断/ユーティリティ

#### GET /healthz
- **認証**: 不要
- **レスポンス**: `200 OK` 固定文字列 `"ok"`

#### GET /readyz
- **認証**: 不要
- **レスポンス**: DB 接続, マイグレーション整合性をチェック

#### GET /config
- **概要**: フロント初期化用設定取得
- **レスポンス**
```json
{
  "auth_providers": ["password"],
  "max_duration_hours": 24,
  "feature_flags": {
    "reports_export": true
  }
}
```

---

## 6. セキュリティ
- **通信**: 本番を想定する場合は HTTPS を推奨。開発時は `http://localhost` を利用する。  
- **入力バリデーション**: Handler 層で構造体バインド + 必須チェック、Usecase 層でドメインルールを検証する。  
- **セッション**: `HttpOnly`・`Secure`（ローカルでは任意）・`SameSite=Lax` の Cookie にユーザーID + 失効時刻 + HMAC 署名を格納し、サーバー側のストレージに依存しない。秘密鍵をローテーションすると全セッションが無効になる。  
- **CORS**: 開発中は `http://localhost:5173` のみ許可。将来のデプロイ時に適宜オリジンを追加する。  
- **ログ**: ログイン失敗や重要イベントはアプリケーションログへ出力し、監査用途には将来対応する。

---

## 7. 非機能要件
- **対象ユーザー**: 個人利用を想定し、同時アクセスは多くても数ユーザー。  
- **レスポンスタイム目標**: 主要 API が 1 秒以内に応答すること。  
- **バックアップ**: SQLite の場合は DB ファイルをコピーしてバックアップし、PostgreSQL を使用する際は `pg_dump` などで取得する。  
- **ログ**: `zap` などの構造化ログは必須ではない。開発段階では標準出力へのテキストログで十分。

---

## 8. 将来拡張

本節の内容は現行 API 仕様には含まれず、今後の検討テーマを列挙しています。

- **OAuth ログイン**: `/api/auth/oauth/{provider}` （Google 等）追加想定
- **Webhooks**: 作業完了通知を外部サービスに POST
- **モバイルクライアント**: 同じ REST API を利用し、必要に応じて軽量なアクセス制御を追加
- **Batch/Worker**: 長時間未終了エントリの自動クローズ、日次集計バッチ

---

## 9. テストポリシー（API 観点）
- **ユニットテスト**: Handler と Usecase に対するテストを `httptest` やインメモリのリポジトリで実施し、バリデーションとエラーパスを確認する。  
- **軽量統合テスト**: SQLite（インメモリ or ファイル）を用意し、`httptest` + 実際の HTTP ハンドラで作業開始〜一覧取得までの最小フローを確認する（将来的に PostgreSQL でも同テストを動かせる設計とする）。  
- **手動確認**: レポート画面など UI からの操作は手動でチェックし、将来的な自動化候補として記録する。

---

## 10. OpenAPI 生成方針
- 現段階では手書きドキュメントのみを維持し、自動生成は導入しない。  
- 将来的に API が安定したら `swaggo/swag` などで `openapi.yaml` を生成し、契約テストへ拡張する。

---

## 11. 変更管理
- API 破壊的変更は `/v2` などバージョンを分けて提供
- 変更時はこの文書と OpenAPI を更新し、`CHANGELOG.md` に追記
- Deprecation はレスポンスヘッダ `Deprecation: true` と `Sunset`（日付）を返す

---

## 12. 参考情報
- 技術設計全体: `docs/DesignDoc.md`
- ER 図: `docs/DesignDoc.md` の「データベース設計」節
- 認証・運用ポリシー: `docs/DesignDoc.md` の「実装方針」「セキュリティ・運用」節
