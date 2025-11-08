# ChronoMe データベース設計書

最終更新: 2025-10-22  
対象バージョン: SQLite 3（初期開発） / PostgreSQL 16（移行先）

---

## 1. 目的と前提
- ChronoMe（一人用タイムカード Web アプリ）の永続化層仕様を定義する。  
- バックエンドは Go（GORM）を想定し、マイグレーションは `golang-migrate` で管理する。  
- 作業時間やレポート機能に必要な主テーブルと制約・インデックス方針を示す。  
- ここで定義した仕様は `docs/DesignDoc.md` および `docs/APIDesign.md` の上位設計と整合する。

### 1.1 運用パラメータ
| 項目 | 値 |
|------|-----|
| 文字コード | UTF-8 |
| タイムゾーン | UTC 固定で保存（アプリ層でローカル変換） |
| ID 生成 | UUID v4（アプリ層/GORM で生成。PostgreSQL 移行時は `gen_random_uuid()` 拡張を使用） |
| 命名規則 | テーブル: `snake_case` 複数形 / カラム: `snake_case` |

### 1.2 並行作業とratio設計方針
- **複数作業の同時実行**: ユーザーは同時に複数の作業を開始可能
- **ratio による時間配分**: 後からガントチャート形式のUIで各作業の時間配分を手動調整
- **自動正規化**: ratio の合計が1.0になるようアプリケーション層で制御

### 1.3 休憩時間（is_break）の除外ロジック
- **識別**: `is_break=true` のエントリは休憩として区別
- **集計除外**: レポート生成時に `WHERE is_break = false` で休憩時間を作業時間から除外
- **UI表示**: タイムライン上では休憩エントリを異なる色（例：グレー）で表示
- **並行作業**: 休憩中でも他の作業を継続可能

### 1.4 タイムゾーン処理の実装詳細
**パターンA: フロントエンド変換方式（推奨）**
- DB保存: 全てUTC (`TIMESTAMP WITH TIME ZONE`)
- API応答: ISO 8601形式 (`2025-10-22T10:00:00Z`)
- フロント変換: ユーザーの `time_zone` 設定とブラウザ `Intl` APIで表示時変換

**パターンB: サーバーサイド変換方式**
- DB保存: UTC統一
- API応答: ユーザーのtime_zone設定に基づいてサーバー側で変換して応答
- フロント: 受信した時刻をそのまま表示

**採用方針**: パターンAを採用（フロントエンド変換）
- 理由: APIがシンプル、キャッシュ効率が良い、異なるタイムゾーンでの閲覧に対応しやすい

- タイムスタンプ列は全テーブルで `created_at`（必須）と、必要に応じて `updated_at` を保持する。

### 1.5 実装ガイドライン
- 論理削除は採用しない（データ削除は物理削除）。必要時はバックアップからリストア。  
- アプリケーションの認可は「ログインユーザー ID と一致するレコードのみ操作可」という前提で設計する。  
- 日付計算ロジック（`duration_sec`, `ratio` チェック等）はユースケース層で実装し、DB 側では最小限のチェック制約を持たせる。

---

## 2. ER 図
```mermaid
erDiagram
    users ||--o{ projects : owns
    users ||--o{ tags : creates
    users ||--o{ entries : records
    projects ||--o{ entries : includes
    entries ||--o{ entry_tags : links
    tags ||--o{ entry_tags : assigns

    users {
        uuid id
        string email
        string password_hash
        string display_name
        string time_zone
        timestamp created_at
        timestamp updated_at
    }

    projects {
        uuid id
        uuid user_id
        string name
        string color
        boolean is_archived
        timestamp created_at
        timestamp updated_at
    }

    tags {
        uuid id
        uuid user_id
        string name
        string color
        timestamp created_at
        timestamp updated_at
    }

    entries {
        uuid id
        uuid user_id
        uuid project_id
        string title
        timestamp started_at
        timestamp ended_at
        integer duration_sec
        boolean is_break
        numeric ratio
        text notes
        timestamp created_at
        timestamp updated_at
    }

    entry_tags {
        uuid entry_id
        uuid tag_id
        timestamp created_at
    }
```

---

## 3. テーブル一覧
| テーブル | 用途 | 主キー | 備考 |
|----------|------|--------|------|
| `users` | 認証・設定を保持するユーザー情報 | `id` | メールアドレスユニーク制約 |
| `projects` | 作業分類（プロジェクト） | `id` | ユーザー内名称ユニーク、アーカイブ可 |
| `tags` | 任意タグ | `id` | ユーザー内名称ユニーク |
| `entries` | 作業時間ログ | `id` | 期間・並行作業割合・休憩フラグを保持 |
| `entry_tags` | エントリとタグの多対多中間 | 複合キー(`entry_id`,`tag_id`) | 作成日時で結び付け履歴を保持 |

---

## 4. テーブル定義詳細

### 4.1 users
| 列名 | 型 | Not Null | 既定値 | 説明 |
|------|----|----------|--------|------|
| `id` | `uuid` | ✅ | `gen_random_uuid()` | ユーザー ID |
| `email` | `citext` | ✅ |  | ログイン用メール（小文字大文字を区別しない） |
| `password_hash` | `text` | ✅ |  | `bcrypt` などでハッシュ化した値 |
| `display_name` | `varchar(50)` |  |  | 画面表示名 |
| `time_zone` | `varchar(40)` | ✅ | `'UTC'` | IANA Time Zone (`Asia/Tokyo` 等) |
| `created_at` | `timestamptz` | ✅ | `now()` | 作成日時 |
| `updated_at` | `timestamptz` | ✅ | `now()` | 更新日時 |

**制約・索引**
- `PRIMARY KEY (id)`
- `UNIQUE (email)`
- `CHECK (display_name IS NULL OR length(display_name) BETWEEN 1 AND 50)`
- `CHECK (time_zone <> '')`

**備考**
- `citext` 拡張を有効にする（`CREATE EXTENSION IF NOT EXISTS citext;`）。  
- `updated_at` はトリガまたはアプリ層で更新する。

### 4.2 projects
| 列名 | 型 | Not Null | 既定値 | 説明 |
|------|----|----------|--------|------|
| `id` | `uuid` | ✅ | `gen_random_uuid()` | プロジェクト ID |
| `user_id` | `uuid` | ✅ |  | 所有ユーザー |
| `name` | `varchar(80)` | ✅ |  | プロジェクト名（ユーザー内ユニーク） |
| `color` | `char(7)` | ✅ | `'#1F2933'` | HEX カラー（`#RRGGBB`） |
| `is_archived` | `boolean` | ✅ | `false` | アーカイブ済みか |
| `created_at` | `timestamptz` | ✅ | `now()` | 作成日時 |
| `updated_at` | `timestamptz` | ✅ | `now()` | 更新日時 |

**制約・索引**
- `PRIMARY KEY (id)`
- `FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE`
- `UNIQUE INDEX idx_projects_user_lower_name ON projects (user_id, lower(name))`
- `CHECK (color ~ '^#[0-9A-Fa-f]{6}$')`
- `INDEX idx_projects_user_id_created_at ON projects(user_id, created_at DESC)`

**備考**
- `is_archived=true` のプロジェクトは API で既定非表示だが、過去のエントリ紐付けは保持する。  
- GORM では `Name` にユニーク制約タグを設定する。

### 4.3 tags
| 列名 | 型 | Not Null | 既定値 | 説明 |
|------|----|----------|--------|------|
| `id` | `uuid` | ✅ | `gen_random_uuid()` |
| `user_id` | `uuid` | ✅ |  | 所有ユーザー |
| `name` | `varchar(40)` | ✅ |  | タグ名（ユーザー内ユニーク） |
| `color` | `char(7)` | ✅ | `'#3B82F6'` | HEX カラー |
| `created_at` | `timestamptz` | ✅ | `now()` |
| `updated_at` | `timestamptz` | ✅ | `now()` |

**制約・索引**
- `PRIMARY KEY (id)`
- `FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE`
- `UNIQUE INDEX idx_tags_user_lower_name ON tags (user_id, lower(name))`
- `CHECK (color ~ '^#[0-9A-Fa-f]{6}$')`
- `INDEX idx_tags_user_id_name ON tags(user_id, name)`

### 4.4 entries
| 列名 | 型 | Not Null | 既定値 | 説明 |
|------|----|----------|--------|------|
| `id` | `uuid` | ✅ | `gen_random_uuid()` |
| `user_id` | `uuid` | ✅ |  | 記録所有者 |
| `project_id` | `uuid` |  |  | 紐付プロジェクト（削除時は `NULL`） |
| `title` | `varchar(120)` | ✅ |  | 作業タイトル |
| `started_at` | `timestamptz` | ✅ |  | 開始時刻 |
| `ended_at` | `timestamptz` |  |  | 終了時刻（進行中は `NULL`） |
| `duration_sec` | `integer` | ✅ | 0 | 所要秒数（アプリ層で算出） |
| `is_break` | `boolean` | ✅ | `false` | 休憩扱いか |
| `ratio` | `numeric(3,2)` | ✅ | 1.00 | 並行作業割合（0.00〜1.00） |
| `notes` | `text` |  |  | 備考 |
| `created_at` | `timestamptz` | ✅ | `now()` |
| `updated_at` | `timestamptz` | ✅ | `now()` |

**制約・索引**
- `PRIMARY KEY (id)`
- `FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE`
- `FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL`
- `CHECK (duration_sec >= 0)`
- `CHECK (ended_at IS NULL OR ended_at >= started_at)`
- `CHECK (ratio >= 0.00 AND ratio <= 1.00)`
- `INDEX idx_entries_user_started_at ON entries(user_id, started_at DESC)`
- `INDEX idx_entries_project_started_at ON entries(project_id, started_at DESC)`
- `INDEX idx_entries_is_break ON entries(user_id, is_break, started_at)`

**備考**
- `duration_sec` は `ended_at - started_at` を元にアプリ側で更新（並行割合を考慮）。  
- 期間検索が多いため `started_at` に DESC の複合インデックスを持たせる。  
- 並行作業割合の整合性（同期間合計 1.0 以下）はユースケース層で検証。

### 4.5 entry_tags
| 列名 | 型 | Not Null | 既定値 | 説明 |
|------|----|----------|--------|------|
| `entry_id` | `uuid` | ✅ |  | 紐付エントリ |
| `tag_id` | `uuid` | ✅ |  | 紐付タグ |
| `created_at` | `timestamptz` | ✅ | `now()` | 紐付日時 |

**制約・索引**
- `PRIMARY KEY (entry_id, tag_id)`
- `FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE`
- `FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE`
- `INDEX idx_entry_tags_tag_id ON entry_tags(tag_id)`

**備考**
- API では `tag_ids` を配列で受け取り差分更新する。  
- タグの多重付与防止のため、複合主キーで重複を禁止する。

---

## 5. ビュー / マテリアライズドビュー（任意提案）
レポート向け集計を高速化するため、以下のビューを検討する。

| 名称 | 種類 | 内容 | 備考 |
|------|------|------|------|
| `vw_entry_daily_summary` | マテリアライズドビュー | `user_id`, `date`, `total_duration_sec`, `billable_duration_sec` を集計 | 日次バッチで再作成 (`REFRESH MATERIALIZED VIEW CONCURRENTLY`) |

> 現段階では必須ではないため、パフォーマンス要件が顕在化したタイミングで導入する。

---

## 6. マイグレーション指針
- 初期マイグレーションでは `CREATE EXTENSION IF NOT EXISTS "pgcrypto";` と `citext` を有効化する。  
- マイグレーションファイルは命名規則 `YYYYMMDDHHMM__description.sql` を用いる。  
- 破壊的変更（カラム削除・型変更）は代替カラム追加 → データ移行 → 旧カラム削除の 3 ステップを徹底する。  
- 外部キー制約に依存する順序でロールバック可能なスクリプトを整備する。

### 6.1 初期化手順（ローカル開発）
1. **SQLite（既定）**  
   - 追加手順は不要。アプリ起動時に `backend/dev.db` が存在しなければ自動作成される。  
   - マイグレーションは GORM の AutoMigrate または `backend/migrations` の SQL を SQLite 向けに適用。
2. **PostgreSQL（移行/検証時）**  
   - Docker Compose 等で PostgreSQL を用意し、`chronome` データベースとユーザーを作成する。  
   - マイグレーションを適用する。  
     ```bash
     golang-migrate -path backend/migrations -database "postgres://user:pass@localhost:5432/chronome?sslmode=disable" up
     ```
   - 適用対象ファイル: `backend/migrations/202510220900__init.sql`

---

## 7. データ保持・バックアップ
- SQLite の場合は `dev.db` のファイルコピーでバックアップを取得する。  
- PostgreSQL の場合は `pg_dump` などで週 1 回程度のバックアップ取得を推奨する。  
- 利用者が自身のアカウントを削除した場合、関連レコードは外部キー `ON DELETE CASCADE` / `SET NULL` 方針に従い削除。  
- GDPR 相当の削除要求発生時は `entries` などの履歴も合わせて削除されることを想定。

---

## 8. 今後の拡張余地

以下はデータベース設計の将来候補であり、現行のテーブル定義には含まれていません。

- `user_settings`（通知やレポート書式）テーブル追加。  
- `entry_attachments`（ファイル添付）テーブル追加。  
- レポート履歴のキャッシュを持つ `report_snapshots` テーブルの導入。

---

本ドキュメントは変更時に Pull Request へ添付し、レビュー結果を反映後にバージョン管理すること。
