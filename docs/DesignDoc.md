# 技術設計書 

## 概要
本ドキュメントは、**ChronoMe（一人用タイムカード Web アプリ）**の技術設計をまとめたものです。  
対象は開発者および提出先の担当者であり、システム構成、データ構造、API設計、開発方針などを示します。

- **目的**：個人の作業時間を記録し、日報・週報・月報として確認できるWebアプリを作成する。
- **構成技術**：  
  - フロントエンド：React + TypeScript  
  - バックエンド：Go（**クリーンアーキテクチャ + GORM + SQLite**）  
  - データベース：SQLite（初期開発） / PostgreSQL（スケール時移行）  
  - 認証：サインド Cookie ベースのセッション  
  - デプロイ対象：ローカル環境（手動実行）

---

## システムアーキテクチャ

```mermaid
graph TD
    subgraph Frontend["フロントエンド (React + TypeScript)"]
        UI["タイマー・日報・週報・月報画面"]
        API["TanStack Query による API 通信"]
    end

    subgraph Backend["バックエンド (Go / Clean Architecture)"]
        Handler["Handler 層 (Chi + REST API)"]
        Usecase["Usecase 層 (業務ロジック)"]
        Repository["Repository 層 (GORM DB操作)"]
        Entity["Entity 層 (ドメインモデル)"]
    end

    DB[(SQLite / PostgreSQL)]

    UI --> API --> Handler
    Handler --> Usecase --> Repository --> DB
    Repository --> Entity
```

上図は現行スコープ（フロントエンドとバックエンド、SQLite を既定としつつ PostgreSQL へ移行可能な構成）を表しています。ファイル保存などの追加要素は将来検討事項であり、現時点の実装対象には含めません。

---

## クリーンアーキテクチャ設計原則

ChronoMe バックエンドはクリーンアーキテクチャを採用し、依存性は常に内向き（Handler → Usecase → Entity）となるよう設計しています。詳細な層ごとの責務、依存性注入、テスト方針、トランザクション設計などは `docs/CleanArchitecture.md` を参照してください。

本ドキュメントでは方針を以下に要約します。

- Usecase 層は Repository Interface や Provider を通して外部リソースへアクセスし、フレームワーク・ORM に直接依存しない。  
- Entity 層は純粋なドメインロジックのみを保持し、副作用を排除する。  
- Interface Adapter 層は DTO 変換や入出力の橋渡しを担い、Infrastructure 層はフレームワーク設定と永続化処理を担当する。  
- 各層でテスト粒度を分け、モック／スタブを活用して保守性・拡張性・テスタビリティを確保する。


---

## 技術スタック

| 分類 | 使用技術 | 補足 |
|------|-----------|------|
| 言語 | Go 1.22.x | バックエンド実装（標準ツールチェーン） |
| アーキテクチャ | **Clean Architecture（簡易構成）** | Handler / Usecase / Repository の3層中心 |
| フロント | React + TypeScript | Vite を利用した SPA |
| ルーティング | **Chi** | 軽量HTTPルータ |
| ORM / DB | **GORM + SQLite（初期） / PostgreSQL（移行時）** | SQLite を既定にし、DSN 切替で PostgreSQL 対応 |
| マイグレーション | 手動スクリプト（必要に応じて `golang-migrate`） | 学習段階では SQL スクリプトを直接適用 |
| 状態管理 | TanStack Query | API通信・キャッシュ管理 |
| テスト | `go test` / React Testing Library | ユニット + 最小限の統合テスト |
| DI | コンストラクタ注入 | 複雑なコンテナは使用しない |
| 開発環境 | Go / npm のローカル実行 | Docker なしで構築 |
| 認証 | サインド Cookie ベースのセッション | 学習用に簡素化 |

---

## ディレクトリ構成（クリーンアーキテクチャ準拠）

```text
backend/
 ├── cmd/
 │   └── main.go                 // DI とアプリケーション起動
 ├── internal/
 │   ├── entity/                 // エンティティ層（最内側）
 │   │   ├── user.go
 │   │   ├── entry.go
 │   │   └── project.go
 │   ├── usecase/                // ユースケース層
 │   │   ├── interfaces.go       // Repository Interface 定義
 │   │   ├── entry_usecase.go
 │   │   ├── auth_usecase.go
 │   │   └── report_usecase.go
 │   ├── repository/             // Infrastructure層（Repository実装）
 │   │   ├── interface.go        // Usecase 層向けポート定義
 │   │   └── gorm/               // GORM による実装
 │   │       ├── entry_repository.go
 │   │       ├── project_repository.go
 │   │       ├── user_repository.go
 │   │       └── models/         // GORM 専用構造体
 │   │           ├── entry.go
 │   │           ├── project.go
 │   │           └── user.go
 │   ├── handler/                // Interface Adapters層
 │   │   ├── auth_handler.go
 │   │   ├── entry_handler.go
 │   │   └── middleware/
 │   └── utils/                  // 共通ユーティリティ
 │       ├── time_provider.go    // テスト可能な時刻取得
 │       └── jwt.go
 ├── migrations/                // SQL スクリプト
 └── go.mod / go.sum

frontend/
 ├── src/
 │   ├── features/               // 機能単位（FSD方式）
 │   ├── components/             // UIコンポーネント
 │   ├── pages/                  // 画面ルーティング
 │   ├── hooks/                  // カスタムフック
 │   └── lib/                    // 共通ユーティリティ
 ├── vite.config.ts
 └── package.json
```

---

## データベース設計（概要）

主要なエンティティとその関係性を以下に示します。  
詳細なテーブル定義、制約、インデックス設計は [DBDesign.md](./DBDesign.md) を参照してください。

```mermaid
erDiagram
    users ||--o{ projects : "owns"
    users ||--o{ entries : "records"
    users ||--o{ tags : "creates"

    projects ||--o{ entries : "has"
    entries ||--o{ entry_tags : "links"
    tags ||--o{ entry_tags : "assigned"

    users {
        uuid id PK
        string email UK
        string password_hash
        string display_name
        string time_zone
        timestamp created_at
        timestamp updated_at
    }

    projects {
        uuid id PK
        uuid user_id FK
        string name
        string color
        boolean is_archived
        timestamp created_at
        timestamp updated_at
    }

    entries {
        uuid id PK
        uuid user_id FK
        uuid project_id FK
        string title
        timestamp started_at
        timestamp ended_at
        integer duration_sec
        boolean is_break
        numeric ratio
        string notes
        timestamp created_at
        timestamp updated_at
    }

    tags {
        uuid id PK
        uuid user_id FK
        string name
        string color
        timestamp created_at
        timestamp updated_at
    }

    entry_tags {
        uuid entry_id FK
        uuid tag_id FK
        timestamp created_at
    }
```

### 主要な設計方針

#### 並行作業の管理
- **複数作業の同時実行**：ユーザーは同時に複数の作業を開始可能
- **ratio による時間配分**：後からガントチャート形式のUIで各作業の時間配分を調整
- **自動計算**：ratio の合計が1.0になるように正規化、実際の作業時間を配分

#### 休憩時間の管理
- **is_break フラグ**：休憩エントリは `is_break=true` で区別
- **集計からの除外**：レポート集計時に休憩時間は作業時間から除外
- **UI での表示**：タイムライン上では休憩時間を異なる色で表示

#### タイムゾーン処理
- **UTC 統一保存**：すべての時刻データは UTC で保存
- **フロントエンド変換**：ユーザーのタイムゾーン設定に基づいて表示時に変換
- **API レスポンス**：ISO 8601 形式（Z接尾辞）で時刻を返却

---

## 認証・セキュリティ設計

| 項目 | 内容 |
|------|------|
| 認証方式 | サインド Cookie ベースのシンプルなセッション |
| トークン保存 | ブラウザの Cookie（HttpOnly, Secure / SameSite=Lax） |
| 暗号化 | bcrypt によるパスワードハッシュ化 |
| 通信 | HTTPS 前提（開発では http://localhost） |
| セッション管理 | ログイン時にセッションIDを発行し、サーバー側（メモリ or DB）で保持 |
| CSRF 対策 | SameSite=Lax 設定で最小限対応 |
| CORS 対策 | 開発時のみ localhost:5173 を許可 |
| パスワードポリシー | 最小8文字、英数字混在（基本的な要件のみ） |

> **補足**  
> 本アプリは学習目的・個人利用を想定し、セキュリティ要件を最小限に抑えています。  
> 本格運用時は OAuth2、Refresh Token、MFA 等の追加実装を推奨します。

---

## API設計（概要）

RESTful API を採用し、JSON でデータを送受信します。  
詳細なリクエスト/レスポンス仕様は [APIDesign.md](./APIDesign.md) を参照してください。

| 区分 | メソッド | パス | 概要 | 認証 |
|------|-----------|------|------|------|
| Auth | POST | /api/auth/signup | 新規登録 | 不要 |
| Auth | POST | /api/auth/login | ログイン | 不要 |
| Auth | POST | /api/auth/logout | ログアウト | 必須 |
| Auth | GET  | /api/auth/me | ログイン中ユーザー情報取得 | 必須 |
| Projects | GET, POST, PATCH, DELETE | /api/projects | プロジェクトCRUD | 必須 |
| Entries | GET, POST, PATCH, DELETE | /api/entries | 作業記録CRUD | 必須 |
| Entries | POST | /api/entries/start | 作業開始 | 必須 |
| Entries | POST | /api/entries/{id}/stop | 作業終了 | 必須 |
| Reports | GET | /api/reports/daily (weekly, monthly) | 集計データ取得 | 必須 |
| Tags | GET, POST, PATCH, DELETE | /api/tags | タグCRUD | 必須 |
| Reports | GET | /reports/weekly | 週次集計データ取得 | 必須 |
| Reports | GET | /reports/monthly | 月次集計データ取得 | 必須 |
| Reports | GET | /reports/export | 集計データのエクスポート | 必須 |

---

## 実装方針

### アーキテクチャ
- **クリーンアーキテクチャ厳守**：依存性ルールを破らない、各層の責務を明確に分離
- **DI パターン**：各コンポーネントはコンストラクタ引数で依存を受け取り、`cmd/server/main.go` で必要な実装を手動で組み立てる
- **Interface 駆動**：テスタビリティを重視した抽象化

### データベース戦略

```mermaid
graph LR
    A[SQLite開発] --> B[機能完成]
    B --> C{スケール必要?}
    C -->|Yes| D[PostgreSQL移行]
    C -->|No| E[SQLite継続]
```

- **結論**: 初期は SQLite を採用し、マイグレーションや抽象化を整備しておくことで、必要になったタイミングで PostgreSQL へ移行できる構成を維持する。
- **実装戦略**
  - SQLite でスタートし、最速で機能開発と学習効果を得る
  - Repository/DSN を抽象化し、ドライバー切替で PostgreSQL へ対応可能にする
  - スケール要件が顕在化したら PostgreSQL ドライバーとマイグレーションを追加実行する
- **具体的アクション**
  1. 初期化: `go mod init chronome`
  2. ドライバー導入（SQLite）: `go get gorm.io/driver/sqlite` / `go get gorm.io/gorm`
  3. DSN 設定でドライバー選択できる構成を実装
  4. 必要時に PostgreSQL ドライバーを追加: `go get gorm.io/driver/postgres`

### データ管理
- **時間ロジック**：新しい作業開始時に未終了エントリを自動終了
- **並行作業対応**：複数作業を許容し、後から `ratio` で割合調整をガントチャート形式のUIで実現
- **休憩管理**：`is_break=true` のエントリは集計レポートから除外、タイムライン上では別色で表示
- **タイムゾーン**：全てUTCで保存し、フロント側でユーザー設定に基づいてローカル表示

### テスト戦略
- **ユニットテスト優先**：Usecase と Entity をインメモリ依存で検証し、バリデーション・ドメインロジックを固める。
- **軽量統合テスト**：インメモリまたはファイル SQLite を用いて Repository の CRUD を確認し、主要制約を網羅する（PostgreSQL でも同テストが動作するように抽象化する）。
- **手動確認の併用**：UI のシナリオは手動で確認し、自動 E2E は将来の課題として切り出す。

詳細な実装ガイドラインは [CleanArchitecture.md](./CleanArchitecture.md) を参照してください。

---

## テスト設計

### テスト方針
- **主要フロー確認**：サインアップ → 記録 → レポート表示の基本シナリオを必ず押さえる。
- **段階的実装**：まずユニットテスト、余力があれば SQLite を使った統合テストを追加し、将来的に PostgreSQL 対応を追加する。
- **実用重視**：テストデータを最小限とし、読みやすさと再利用性を重視する。

### テスト種別と実行環境
- **Unit（Usecase/Entity）**：`go test ./internal/...` - インメモリスタブを利用
- **Integration（Repository）**：`go test ./internal/repository/...` - SQLite（インメモリ or ファイル）を利用
- **E2E（手動確認）**：ブラウザを用いた主要フローの手動チェック（自動化は将来検討）

### 時刻依存テストの対応
実装コストを抑えつつテスタビリティを確保するため、以下を採用：
- Usecase 層には `TimeProvider` インターフェース（`Now() time.Time` のみ）を注入する
- 本番ではシステム時刻実装、テストでは固定時刻を返すスタブを用意する
- DI 経由で差し替え、時刻依存ロジックを安定して検証できるようにする

詳細なテスト実装例は [TestStrategy.md](./TestStrategy.md) を参照してください。

### Mock 生成
```bash
# mockery を使用してテスト用 Mock を自動生成
mockery --dir=internal/usecase --name=EntryRepository --output=mocks
mockery --dir=internal/usecase --name=UserRepository --output=mocks
```

### 最小ディレクトリ
```
backend/
├── cmd/
├── internal/
│   ├── handler/
│   ├── usecase/
│   ├── repository/
│   └── entity/
├── mocks/              # （任意）テスト用モック
├── migrations/
└── dev.db              # 開発用 SQLite（自動生成）。`.env` で DSN を切り替え可能

frontend/
├── src/
└── package.json
```

### 実行例
```bash
# Unit Test
go test ./internal/... -v

# （任意）PostgreSQL を使った軽量な統合テスト
TEST_DATABASE_URL=postgres://chronome_test:chronome_test@localhost:5433/chronome_test?sslmode=disable go test ./internal/repository/... -v
```

### 代表シナリオ（E2E）
1) **Auth**：signup → login → me（不正認証は 401）  
2) **Entries**：start → stop → list（未終了があれば自動終了）  
3) **Reports**：`GET /reports/daily` で日次集計取得（UTC保存／JST表示の簡易確認）

> メモ：時刻依存を避けるため、Usecaseに `TimeProvider` Interface を注入し、テスト時は固定時刻を返すスタブを使用する。

---

## 開発環境構築手順（ローカル実行）

1. `git clone <repo>`
2. `cd backend && go mod download`
3. `cd frontend && npm install`
4. バックエンド: `go run ./cmd/server`
5. フロントエンド: `npm run dev`
6. SQLite の場合は初回起動時に `dev.db` が自動生成される。PostgreSQL で運用する際は `backend/migrations` の SQL を適用する。

---

## 今後の拡張案（参考）

本節は現行リリースの範囲外とし、今後着手する可能性のあるアイデアを整理したものです。

- PWA化によるオフライン対応
- 通知機能（Push API）
- モバイルアプリ（React Native）

---

## まとめ

ChronoMe は、**Go + React + SQLite（初期） / PostgreSQL（移行可能）** を用いた学習目的のタイムカードWebアプリです。  
**クリーンアーキテクチャを厳守**し、テスタブルで保守性の高いコードベースを構築します。  
現段階ではローカル開発・学校提出を主目的とし、将来的な拡張（モバイル対応・クラウド連携）も想定しています。
