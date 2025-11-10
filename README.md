# ChronoMe
2025/11 個人制作
# ChronoMe

> 一人用タイムカード Web アプリケーション  
> 2025/11 個人制作

[![Go](https://img.shields.io/badge/Go-1.22-blue.svg)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18-blue.svg)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-blue.svg)](https://www.typescriptlang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue.svg)](https://www.postgresql.org/)

## 概要

ChronoMe は、個人の作業時間を記録・管理するためのタイムカード Web アプリケーションです。  
**クリーンアーキテクチャ**を採用し、保守性とテスタビリティを重視して開発されています。

### 主な機能

- ⏱️ **タイムトラッキング**: 作業開始/終了の記録
- 📊 **レポート機能**: 日次・週次・月次の作業時間集計
- 🏷️ **プロジェクト管理**: 作業をプロジェクト別に分類
- 🔖 **タグ機能**: 作業内容の詳細分類
- 📱 **レスポンシブ対応**: モバイル・デスクトップ両対応
- 🔐 **セッション認証**: サインド Cookie を用いたシンプルなユーザー認証

### 技術スタック

#### フロントエンド
- **React 18** + **TypeScript**
- **Vite** (開発サーバー/ビルド)
- **TanStack Query** (API フェッチ・キャッシュ)
- **Tailwind CSS** (スタイリング)

#### バックエンド
- **Go 1.22** + **シンプルな Clean Architecture 構成**
- **Chi** (HTTP ルーター)
- **GORM** (ORM)
- **SQLite（初期開発） / PostgreSQL（スケール時移行）**

#### 開発ツール
- Go / npm の標準ツールチェーン
- `go test`, `npm test` による軽量テスト

---

## アーキテクチャ

```mermaid
graph TD
    subgraph "Frontend (React + TypeScript)"
        A[UI Components]
        B[TanStack Query]
        C[Custom Hooks]
    end
    
    subgraph "Backend (Go + Clean Architecture)"
        D[Handler Layer]
        E[Usecase Layer]
        F[Repository Interface]
        G[Entity Layer]
        H[Repository Implementation]
    end
    
    I[(PostgreSQL)]
    
    A --> B
    B --> D
    D --> E
    E --> F
    F --> G
    H --> F
    H --> I
```

### クリーンアーキテクチャの採用

- **Entity層**: ビジネスルールとドメインモデル
- **Usecase層**: アプリケーション固有のビジネスロジック
- **Repository Interface**: データアクセスの抽象化
- **Handler層**: HTTP リクエスト/レスポンスの処理

詳細は [技術設計書](./docs/DesignDoc.md) を参照してください。

---

## データベース戦略

```mermaid
graph LR
    A[SQLite開発] --> B[機能完成]
    B --> C{スケール必要?}
    C -->|Yes| D[PostgreSQL移行]
    C -->|No| E[SQLite継続]
```

- 初期は SQLite + GORM で最速の機能開発に集中します。
- Repository 層は DSN とドライバーを抽象化し、`gorm.io/driver/sqlite` / `gorm.io/driver/postgres` を切り替え可能にします。
- スケール要件が発生した際は PostgreSQL ドライバーを追加し、マイグレーションを適用して移行します。

```bash
# 初期セットアップ
go mod init chronome
go get gorm.io/gorm gorm.io/driver/sqlite

# 移行時に PostgreSQL ドライバーを追加
go get gorm.io/driver/postgres
```

---

## 開発環境構築

### 前提条件

- Node.js 18+（フロントエンド開発用）
- Go 1.22+（バックエンド開発用）
- SQLite 3（同梱バイナリで利用可能、追加セットアップ不要）
- ※ スケール検討時に PostgreSQL 16 へ移行できるよう設計

### 1. リポジトリのクローン

```bash
git clone https://github.com/yourusername/ChronoMe.git
cd ChronoMe
```

### 2. 依存関係のインストール

```bash
cd backend
go mod download

cd ../frontend
npm install
```

### 3. 開発サーバーの起動

```bash
# バックエンド
cd backend
go run ./cmd/server

# フロントエンド（別ターミナルで）
cd frontend
npm run dev
```

バックエンドはローカルの SQLite ファイル（既定: `backend/dev.db`）を使用します。`.env` で DSN を差し替えることで PostgreSQL へ移行できます。

### 4. 代表的な環境変数

| 変数名 | 役割 | 既定値 |
| --- | --- | --- |
| `APP_ENV` | `production` の場合は安全設定が強制されます。 | `development` |
| `SESSION_SECRET` | サインド Cookie の暗号鍵。マルチインスタンス運用では必ず安全な値に置き換えてください。 | `dev-secret-change-me` |
| `SESSION_COOKIE_SECURE` | Cookie に `Secure` フラグを付与するかどうか | `APP_ENV` が `production` のとき `true` |
| `SESSION_TTL` | サインド Cookie の有効期間 (Go duration) | `12h` |
| `DEFAULT_PROJECT_COLOR` | 新規プロジェクトのデフォルト HEX カラー | `#3B82F6` |
| `SERVER_ADDRESS` | HTTP サーバーのバインド先 | `:8080` |
| `ALLOWED_ORIGIN` | CORS 許可ドメイン | `http://localhost:5173` |
| `VITE_BACKEND_URL` | (Front) Vite の `/api` プロキシ先 | `http://localhost:8080` |

サインド Cookie は `SESSION_TTL`（既定 12h）を過ぎると自動で失効します。`SESSION_SECRET` をローテーションすると既存 Cookie も検証に失敗するため、強制ログアウトさせたい場合は鍵を変更してください。ログアウト時は `Max-Age=-1` のクッキーを返し、クライアント側でも即時無効化されます。
## ローカル開発

### バックエンド開発

```bash
cd backend

# 開発サーバー起動
go run ./cmd/server
```

### フロントエンド開発

```bash
cd frontend

# 開発サーバー起動
npm run dev
```

> 開発中のフロントエンドは Vite のプロキシ機能で `/api` へのリクエストを `VITE_BACKEND_URL`（既定 `http://localhost:8080`）へ転送します。バックエンドポートを変更した場合は `.env` でこの値を上書きしてください。

---

## テスト

### バックエンドテスト

```bash
cd backend

# ユニットテスト
go test ./internal/usecase/... -v

# （任意）PostgreSQL を用いた統合テスト
TEST_DATABASE_URL=postgres://chronome_test:chronome_test@localhost:5433/chronome_test?sslmode=disable go test ./internal/adapter/db/... -v
```

### フロントエンドテスト

```bash
cd frontend

# ユニットテスト（Jest + React Testing Library想定）
npm run test
```

---

## API ドキュメント

### 認証 API

| メソッド | エンドポイント | 説明 | 認証 |
|----------|----------------|------|------|
| POST | `/api/auth/signup` | ユーザー登録 | 不要 |
| POST | `/api/auth/login` | ログイン | 不要 |
| POST | `/api/auth/logout` | ログアウト | 必須 |
| GET | `/api/auth/me` | ユーザー情報取得 | 必須 |

### エントリ API

| メソッド | エンドポイント | 説明 | 認証 |
|----------|----------------|------|------|
| GET | `/api/entries` | エントリ一覧取得 | 必須 |
| POST | `/api/entries` | エントリ作成 | 必須 |
| PATCH | `/api/entries/:id` | エントリ更新 | 必須 |
| DELETE | `/api/entries/:id` | エントリ削除 | 必須 |

### レポート API

| メソッド | エンドポイント | 説明 | 認証 |
|----------|----------------|------|------|
| GET | `/api/reports/daily` | 日次レポート | 必須 |
| GET | `/api/reports/weekly` | 週次レポート | 必須 |
| GET | `/api/reports/monthly` | 月次レポート | 必須 |

詳細なAPIドキュメントは [API設計書](./docs/APIDesign.md) を参照してください。

---

## ディレクトリ構成

```
ChronoMe/
├── backend/                 # Go バックエンド
│   ├── cmd/server           # アプリケーションエントリーポイント
│   ├── internal/
│   │   ├── domain/         # エンティティ & Repository インターフェース
│   │   ├── usecase/        # ユースケース + dto/provider
│   │   └── adapter/        # HTTP ハンドラ、GORM 実装、config/session/time などのインフラ
│   └── test/fakes          # ユースケース/ハンドラ向けフェイク
├── frontend/               # React フロントエンド
│   ├── src/
│   │   ├── components/     # UIコンポーネント
│   │   ├── features/       # 機能別モジュール
│   │   ├── pages/          # 画面ルーティング
│   │   ├── hooks/          # カスタムフック
│   │   └── lib/           # ユーティリティ
│   └── public/
├── docs/                   # ドキュメント
└── README.md
```

---

## 使用方法

### 1. ユーザー登録・ログイン

1. http://localhost:5173 にアクセス
2. 「新規登録」からアカウントを作成
3. ログイン画面でメールアドレスとパスワードを入力

### 2. プロジェクトの作成

1. ダッシュボードから「プロジェクト管理」を選択
2. プロジェクト名と色を設定して作成

### 3. 作業時間の記録

1. ダッシュボードの「作業開始」ボタンをクリック
2. 作業タイトルとプロジェクトを選択
3. 作業終了時に「作業終了」ボタンをクリック

### 4. レポートの確認

1. サイドバーから「レポート」を選択
2. 日次・週次・月次のタブで期間を切り替え
3. グラフとテーブルで作業時間を確認

---

## ライセンス

このプロジェクトは個人制作物です。学習目的での使用は自由ですが、商用利用の際は事前にご連絡ください。

---

## 開発者

**Miyu** - 2025年11月  
フルスタック Web アプリケーション開発の学習プロジェクト

---

## 今後の予定

- [ ] PWA対応（オフライン機能）
- [ ] 通知機能（Web Push API）
- [ ] CSV/PDF エクスポート
- [ ] モバイルアプリ（React Native）
- [ ] チーム機能（複数ユーザー対応）

---

## 参考資料

- [技術設計書](./docs/DesignDoc.md)
- [API設計書](./docs/APIDesign.md)
- [データベース設計書](./docs/DBDesign.md)
- [クリーンアーキテクチャ実装ガイド](./docs/CleanArchitecture.md)
- [テスト戦略書](./docs/TestStrategy.md)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
