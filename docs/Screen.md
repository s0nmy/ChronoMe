s0nmy: これらのドキュメントの矛盾点や指摘事項、気になる点をすべて列挙してください

GitHub Copilot: これらのドキュメントを詳細に分析した結果、以下の矛盾点や指摘事項、気になる点を発見しました：

## 1. 技術スタック・依存関係の矛盾

### 1.1 データベース関連
- **DesignDoc.md**: 「開発時は SQLite、将来的に PostgreSQL」
- **README.md**: 「PostgreSQL 16」バッジ表示、「開発中 SQLite を既定」
- **DBDesign.md**: 「PostgreSQL 16」を対象、`citext`拡張やPostgreSQL固有機能を使用
- **矛盾**: SQLite開発なのにPostgreSQL固有機能（`citext`, `gen_random_uuid()`）を設計に含んでいる

### 1.2 マイグレーション管理
- **DesignDoc.md**: 「手動スクリプト（必要に応じて `golang-migrate`）」
- **DBDesign.md**: 「マイグレーションは `golang-migrate` で管理する」
- **矛盾**: マイグレーション戦略が一貫していない

## 2. API設計の不整合

### 2.1 エンドポイント定義
- **DesignDoc.md**: `/api/entries/start`, `/api/entries/{id}/stop`
- **APIDesign.md**: `/entries/start`, `/entries/{entry_id}/stop`
- **矛盾**: `/api`プレフィックスの有無が不一致

### 2.2 並行作業の扱い
- **DesignDoc.md**: 「新しい作業開始時に未終了エントリを自動終了」
- **APIDesign.md**: 「未終了エントリがある場合でも新規作成を許可（並行作業対応）」
- **矛盾**: 並行作業を許可するかどうかで相反する記述

## 3. クリーンアーキテクチャ実装の矛盾

### 3.1 ディレクトリ構成
- **DesignDoc.md**: `internal/repository/gorm_user_repository.go`
- **CleanArchitecture.md**: `internal/repository/gorm/entry_repository.go`
- **不整合**: ディレクトリ構成が異なる

### 3.2 依存性注入
- **DesignDoc.md**: 「コンストラクタ注入によるシンプルな依存性管理」
- **CleanArchitecture.md**: 「DI コンテナ的な役割を担う」
- **曖昧**: 具体的なDI実装方法が不明確

## 4. データモデル設計の問題

### 4.1 時刻管理
- **DBDesign.md**: 「DB保存: 全てUTC (`TIMESTAMP WITH TIME ZONE`)」
- **実装**: SQLiteは`TIMESTAMP WITH TIME ZONE`をサポートしない
- **技術的矛盾**: 開発環境で使用するSQLiteで実現不可能

### 4.2 UUIDの生成
- **DBDesign.md**: 「ID 生成: UUID v4（`gen_random_uuid()` を利用）」
- **実装**: SQLiteは`gen_random_uuid()`をサポートしない
- **技術的矛盾**: PostgreSQL固有機能をSQLite環境で使用不可

## 5. テスト戦略の矛盾

### 5.1 モック使用方針
- **TestStrategy.md**: 「Mock最小化」「シンプルなスタブやインメモリ実装で検証」
- **CleanArchitecture.md**: 「testify/mock などを利用する」
- **DesignDoc.md**: Mock生成コマンド（mockery）の記載
- **矛盾**: Mock使用方針が一貫していない

### 5.2 テスト実行環境
- **TestStrategy.md**: 「SQLite（インメモリ or ファイル）を利用」
- **DesignDoc.md**: 「SQLite を使った軽量な統合テスト」
- **実装**: GORMでSQLiteインメモリ構文が複雑（`file::memory:?cache=shared`）

## 6. セキュリティ設計の課題

### 6.1 認証方式の詳細
- **DesignDoc.md**: 「サインド Cookie ベースのセッション」
- **APIDesign.md**: 「サインド Cookie ベースのセッション」
- **不明確**: サインド Cookieの具体的な実装方法（JWT？HMAC？）が未定義

### 6.2 CSRF対策
- **DesignDoc.md**: 「SameSite=Lax 設定で最小限対応」
- **APIDesign.md**: 同様の記載
- **不十分**: CSRFトークンなしでSameSiteのみは不十分

## 7. API仕様の詳細不足

### 7.1 エラーレスポンス
- **APIDesign.md**: エラーレスポンス形式を定義
- **実装**: 各エンドポイントでの具体的なエラーコードが不完全

### 7.2 バリデーション詳細
- **APIDesign.md**: 基本的なバリデーションルールのみ
- **実装**: 詳細な制約（文字数、形式等）が不明確

## 8. パフォーマンス・運用面の課題

### 8.1 インデックス設計
- **DBDesign.md**: 基本的なインデックスのみ定義
- **実装**: 複合クエリ（期間+プロジェクト+タグ）のインデックス最適化が不足

### 8.2 ページネーション
- **APIDesign.md**: 「最大 per_page: 100」
- **実装**: 大量データ時のパフォーマンス考慮が不足

## 9. バージョニング・互換性

### 9.1 API バージョニング
- **APIDesign.md**: 「ベース URL: `https://api.chronome.example.com/v1`」
- **実装**: 実際のエンドポイントでバージョンが不明確

### 9.2 データベースマイグレーション
- **DBDesign.md**: マイグレーション方針を記載
- **実装**: ロールバック戦略が不明確

## 10. ドキュメント整合性

### 10.1 更新日時
- **DesignDoc.md**: 「2025年10月23日」
- **他ドキュメント**: 異なる日付（10月16日、10月22日等）
- **問題**: ドキュメント間で更新日時が不整合

### 10.2 参照関係
- 各ドキュメントが相互参照しているが、実際の内容に矛盾が存在
- 特にクリーンアーキテクチャの実装詳細で不整合

## 推奨改善策

1. **技術スタック統一**: SQLite開発ならPostgreSQL固有機能を削除、またはPostgreSQL統一
2. **API仕様統一**: エンドポイント、動作仕様の記載を全ドキュメントで整合させる
3. **クリーンアーキテクチャ詳細化**: 具体的な実装方法とディレクトリ構成を統一
4. **テスト戦略明確化**: Mock使用方針を決定し、一貫した戦略を採用
5. **セキュリティ強化**: CSRF対策やセッション管理の具体的実装を明確化
6. **ドキュメント同期**: 全ドキュメントの更新日時と内容を同期

これらの矛盾を解決することで、より実装しやすく保守性の高い設計になると考えられます。

s0nmy: SQLiteを実装した後のPostgreSQLへの変更はどのような手順でしょうか？また、実装ハードルはどうでしょうか？

GitHub Copilot: SQLiteからPostgreSQLへの移行手順と実装ハードルについて詳しく説明します。

## 移行手順

### 1. データベース接続層の抽象化（事前準備）

```go
// internal/infrastructure/database/connection.go
type DatabaseConfig struct {
    Driver   string // "sqlite" or "postgres"
    DSN      string
    // PostgreSQL用の追加設定
    MaxOpenConns int
    MaxIdleConns int
}

func NewDatabase(config DatabaseConfig) (*gorm.DB, error) {
    switch config.Driver {
    case "sqlite":
        return gorm.Open(sqlite.Open(config.DSN), &gorm.Config{})
    case "postgres":
        return gorm.Open(postgres.Open(config.DSN), &gorm.Config{})
    default:
        return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
    }
}
```

### 2. GORM Modelの調整

```go
// 現在のSQLite用モデル
type User struct {
    ID        string    `gorm:"type:text;primaryKey;default:uuid_generate_v4()" json:"id"`
    Email     string    `gorm:"type:text;uniqueIndex" json:"email"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// PostgreSQL対応版
type User struct {
    ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Email     string    `gorm:"type:citext;uniqueIndex" json:"email"`  // citext拡張
    CreatedAt time.Time `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
}
```

### 3. マイグレーション戦略

#### フェーズ1: デュアル対応実装
```go
// migrations/sqlite/001_init.sql
CREATE TABLE users (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
    email TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

// migrations/postgres/001_init.sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### フェーズ2: 設定による切り替え
```go
// cmd/server/main.go
func main() {
    // 環境変数または設定ファイルで切り替え
    dbDriver := os.Getenv("DB_DRIVER") // "sqlite" or "postgres"
    
    var config DatabaseConfig
    switch dbDriver {
    case "postgres":
        config = DatabaseConfig{
            Driver: "postgres",
            DSN:    os.Getenv("POSTGRES_DSN"),
        }
    default:
        config = DatabaseConfig{
            Driver: "sqlite",
            DSN:    "./dev.db",
        }
    }
    
    db, err := NewDatabase(config)
    // ...
}
```

### 4. データ移行手順

#### 手動データ移行（小規模データ）
```go
// tools/migrate_sqlite_to_postgres.go
func MigrateData(sqliteDB, postgresDB *gorm.DB) error {
    // 1. SQLiteからデータを取得
    var users []User
    if err := sqliteDB.Find(&users).Error; err != nil {
        return err
    }
    
    // 2. PostgreSQLに挿入（バッチ処理）
    batchSize := 1000
    for i := 0; i < len(users); i += batchSize {
        end := i + batchSize
        if end > len(users) {
            end = len(users)
        }
        
        if err := postgresDB.CreateInBatches(users[i:end], batchSize).Error; err != nil {
            return err
        }
    }
    
    return nil
}
```

#### CSVエクスポート/インポート（大規模データ）
```bash
# SQLiteからCSVエクスポート
sqlite3 dev.db -header -csv "SELECT * FROM users;" > users.csv

# PostgreSQLにインポート
psql -d chronome -c "\COPY users FROM 'users.csv' WITH CSV HEADER;"
```

### 5. 段階的移行実装

```go
// internal/repository/database_adapter.go
type DatabaseAdapter struct {
    sqlite   *gorm.DB
    postgres *gorm.DB
    usePostgres bool
}

func (da *DatabaseAdapter) GetDB() *gorm.DB {
    if da.usePostgres && da.postgres != nil {
        return da.postgres
    }
    return da.sqlite
}

// 段階的切り替え（読み取り専用から開始）
func (da *DatabaseAdapter) GetReadDB() *gorm.DB {
    // 読み取りはPostgreSQL、書き込みはSQLiteのような段階的移行も可能
    return da.GetDB()
}
```

## 実装ハードル分析

### 🟢 低ハードル項目

#### 1. GORM基本操作
```go
// SQLite、PostgreSQL共通で動作
var users []User
db.Where("created_at > ?", time.Now().AddDate(0, 0, -7)).Find(&users)
```

#### 2. 基本的なCRUD操作
- GORMの抽象化により、基本操作はほぼ変更不要
- 設定ファイルの変更のみで切り替え可能

#### 3. トランザクション処理
```go
// 共通コード
func (r *repository) WithTransaction(fn func(*gorm.DB) error) error {
    return r.db.Transaction(fn)
}
```

### 🟡 中ハードル項目

#### 1. UUID生成方法の統一
```go
// 統一されたUUID生成
func GenerateUUID() string {
    return uuid.New().String()
}

// GORM Hookで対応
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == "" {
        u.ID = GenerateUUID()
    }
    return nil
}
```

#### 2. タイムゾーン処理
```go
// 統一されたタイムゾーン処理
func (r *repository) parseTime(t time.Time) time.Time {
    // PostgreSQLの場合はそのまま、SQLiteの場合はUTC変換
    if r.isPostgreSQL {
        return t
    }
    return t.UTC()
}
```

#### 3. マイグレーション管理
```bash
# golang-migrate導入
go install -tags 'postgres sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# PostgreSQL用マイグレーション実行
migrate -path migrations/postgres -database "postgres://user:pass@localhost/chronome?sslmode=disable" up
```

### 🔴 高ハードル項目

#### 1. データ型の非互換性

| 項目 | SQLite | PostgreSQL | 対策 |
|------|--------|------------|------|
| UUID | TEXT | UUID + pgcrypto拡張 | アプリ側で生成統一 |
| 大文字小文字 | COLLATE NOCASE | CITEXT拡張 | 拡張使用or正規化処理 |
| BOOLEAN | INTEGER(0/1) | BOOLEAN | GORM変換に依存 |
| DATETIME | TEXT/INTEGER | TIMESTAMPTZ | 時刻処理ライブラリ統一 |

#### 2. SQL機能の差異
```sql
-- SQLite（制限あり）
SELECT * FROM entries 
WHERE date(started_at) = date('now');

-- PostgreSQL（豊富な関数）
SELECT * FROM entries 
WHERE started_at::date = CURRENT_DATE;
```

#### 3. パフォーマンス特性の違い
```go
// SQLiteは単一書き込み、PostgreSQLは並行処理対応
func (r *repository) optimizeForDB() {
    if r.isPostgreSQL {
        // 接続プール設定
        r.db.DB().SetMaxOpenConns(25)
        r.db.DB().SetMaxIdleConns(5)
    } else {
        // SQLiteは接続プール不要
        r.db.DB().SetMaxOpenConns(1)
    }
}
```

## 推奨移行戦略

### 段階的移行アプローチ

```mermaid
graph LR
    A[SQLite開発] --> B[デュアル対応実装]
    B --> C[PostgreSQL並行運用]
    C --> D[PostgreSQL完全移行]
    D --> E[SQLite削除]
```

#### Phase 1: 準備期間（1-2週間）
- データベース接続層の抽象化
- 環境変数による切り替え機能実装
- PostgreSQL用マイグレーション作成

#### Phase 2: 並行運用（1週間）
- PostgreSQL環境構築
- データ移行スクリプト実行
- 両環境での動作確認

#### Phase 3: 完全移行（数日）
- PostgreSQL環境での本格運用開始
- SQLite関連コード削除
- ドキュメント更新

### コスト見積もり

| 項目 | 工数 | 難易度 |
|------|------|--------|
| 接続層抽象化 | 0.5人日 | 低 |
| マイグレーション作成 | 1人日 | 中 |
| データ移行スクリプト | 0.5人日 | 中 |
| 型互換性対応 | 1人日 | 高 |
| 統合テスト | 1人日 | 中 |
| **合計** | **4人日** | **中** |

## 結論

**実装ハードル: 中程度**

- **利点**: GORMによる抽象化で基本操作は変更最小限
- **課題**: データ型互換性とSQL方言の違いに注意が必要
- **推奨**: 段階的移行により安全に実施可能

現在のクリーンアーキテクチャ設計により、Repository層の実装変更のみで対応でき、Usecase層以上は影響を最小限に抑えられます。

2 種類のライセンスで類似のコードが見つかりました

s0nmy: 元からPostgreSQLにするかSQLiteにするか悩んでいます
メリット・デメリットを教えて下さい

GitHub Copilot: ChronoMeプロジェクトにおけるPostgreSQLとSQLiteの選択について、詳細な比較分析を提供します。

## 🗃️ PostgreSQL vs SQLite 比較分析

### SQLite のメリット・デメリット

#### ✅ SQLiteのメリット

**1. 開発効率・学習コスト**
```bash
# セットアップが超簡単
go run ./cmd/server  # これだけでDB込みで起動
```
- **ゼロコンフィグ**: インストール・設定不要
- **ファイルベース**: `dev.db`ファイル1つで完結
- **デバッグ容易**: SQLiteブラウザで直接データ確認可能
- **バックアップ簡単**: ファイルコピーのみ

**2. 学習・プロトタイピング適性**
- **軽量**: メモリ使用量最小限
- **高速**: 単純なクエリは非常に高速
- **ポータブル**: 環境依存なし、どこでも動作
- **テスト向け**: インメモリDBでテスト高速化

**3. デプロイメント**
```go
// 単一バイナリでDB込み配布可能
go build -o chronome ./cmd/server
./chronome  // これだけで全て動作
```

#### ❌ SQLiteのデメリット

**1. 機能制限**
```sql
-- SQLiteで使えない機能の例
-- 1. UUIDの自動生成
CREATE TABLE users (
    id TEXT PRIMARY KEY, -- UUIDを手動生成する必要
    email TEXT
);

-- 2. 大文字小文字区別なし検索が複雑
-- PostgreSQL: email CITEXT
-- SQLite: COLLATE NOCASEや関数を使用

-- 3. 制約の制限
-- 外部キー制約の変更が困難
-- CHECK制約の複雑な表現に制限
```

**2. データ型の問題**
```go
// SQLiteの型変換課題
type User struct {
    ID        string    `gorm:"type:text"`           // UUID as TEXT
    CreatedAt time.Time `gorm:"type:datetime"`       // タイムゾーン情報なし
    IsActive  bool      `gorm:"type:integer"`        // boolean as 0/1
}
```

**3. スケーラビリティ制限**
- **並行書き込み**: 1つのプロセスのみ書き込み可能
- **レプリケーション**: なし
- **パーティショニング**: なし
- **インデックス**: 基本的なもののみ

### PostgreSQL のメリット・デメリット

#### ✅ PostgreSQLのメリット

**1. 本格的なRDBMS機能**
```sql
-- PostgreSQLの強力な機能
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT UNIQUE NOT NULL,  -- 大文字小文字区別なし
    created_at TIMESTAMPTZ DEFAULT NOW()  -- タイムゾーン対応
);

-- 複雑な集計クエリ
SELECT 
    date_trunc('week', started_at) as week,
    sum(duration_sec) as total_seconds,
    array_agg(DISTINCT project_id) as projects
FROM entries 
WHERE user_id = $1 
GROUP BY week 
ORDER BY week;
```

**2. データ型の豊富さ**
```go
type Entry struct {
    ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid()"`
    StartedAt time.Time  `gorm:"type:timestamptz"`
    Metadata  datatypes.JSON `gorm:"type:jsonb"`  // JSON型ネイティブサポート
}
```

**3. パフォーマンス・スケーラビリティ**
- **並行処理**: 複数クライアントの同時書き込み
- **インデックス**: 部分インデックス、関数インデックス等
- **クエリ最適化**: 高度なオプティマイザ
- **レプリケーション**: マスター・スレーブ構成可能

#### ❌ PostgreSQLのデメリット

**1. 環境構築の複雑さ**
```bash
# PostgreSQL セットアップが必要
# macOS
brew install postgresql
brew services start postgresql
createdb chronome

# 環境変数設定
export DATABASE_URL="postgres://user:pass@localhost/chronome"

# マイグレーション実行
migrate -path migrations -database $DATABASE_URL up
```

**2. 学習コスト**
- **設定ファイル**: postgresql.conf, pg_hba.conf等
- **ユーザー管理**: ロール・権限設定
- **メンテナンス**: VACUUM, ANALYZE等の運用知識
- **トラブルシューティング**: ログ解析、パフォーマンス調査

**3. リソース使用量**
```yaml
# 最小構成でも一定のリソース消費
shared_buffers: 128MB
effective_cache_size: 4GB
max_connections: 100
```

## 🎯 プロジェクト特性による選択指針

### ChronoMeの要件分析

#### 現在の要件
- **ユーザー数**: 1名（個人用）
- **データ量**: 年間数千レコード程度
- **クエリ複雑度**: 基本的なCRUD + 集計
- **開発期間**: 学習目的（短期）
- **運用**: ローカル環境

#### 将来の可能性
- **スケールアップ**: 複数ユーザー対応
- **機能拡張**: 高度なレポート機能
- **運用**: クラウドデプロイ

### 🏆 推奨判断マトリクス

| 優先度 | 要素 | SQLite | PostgreSQL |
|--------|------|--------|------------|
| 🔴 高 | 開発速度 | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| 🔴 高 | 学習コスト | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| 🟡 中 | 機能豊富さ | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| 🟡 中 | 拡張性 | ⭐ | ⭐⭐⭐⭐⭐ |
| 🟢 低 | 本番運用性 | ⭐⭐ | ⭐⭐⭐⭐⭐ |

## 💡 最終推奨案

### パターン1: SQLite優先（推奨）

```go
// 設定による切り替え設計
type Config struct {
    Database struct {
        Driver string `default:"sqlite"`
        DSN    string `default:"./dev.db"`
    }
}

func NewDatabase(config Config) (*gorm.DB, error) {
    switch config.Database.Driver {
    case "postgres":
        return gorm.Open(postgres.Open(config.Database.DSN), &gorm.Config{})
    default:
        return gorm.Open(sqlite.Open(config.Database.DSN), &gorm.Config{})
    }
}
```

**理由**
1. **学習効率最大化**: 環境構築で時間を取られない
2. **開発サイクル高速化**: すぐにコードに集中できる
3. **テスト容易性**: インメモリテストが簡単
4. **デプロイ簡素化**: 単一バイナリ配布

**制約対応策**
```go
// UUID生成の統一
func generateUUID() string {
    return uuid.New().String()
}

// タイムゾーン処理の統一
func (r *repository) parseTime(t time.Time) time.Time {
    return t.UTC()
}

// 大文字小文字検索の統一
func (r *repository) findByEmailIgnoreCase(email string) (*User, error) {
    var user User
    return &user, r.db.Where("LOWER(email) = LOWER(?)", email).First(&user).Error
}
```

### パターン2: ハイブリッド設計

```go
// 段階的移行を前提とした設計
type DatabaseInterface interface {
    Create(interface{}) error
    Find(interface{}, ...interface{}) error
    // 共通インターフェース
}

type SQLiteRepository struct {
    db *gorm.DB
}

type PostgresRepository struct {
    db *gorm.DB
}

// 両方とも同じインターフェースを実装
```

**移行タイミング**
1. **Phase 1**: SQLiteで機能実装完了
2. **Phase 2**: PostgreSQL対応を追加
3. **Phase 3**: 必要に応じて移行

## 🎯 結論と推奨アクション

### 最終推奨: **SQLite First + Migration Ready**

```mermaid
graph LR
    A[SQLite開発] --> B[機能完成]
    B --> C{スケール必要?}
    C -->|Yes| D[PostgreSQL移行]
    C -->|No| E[SQLite継続]
```

### 実装戦略
1. **SQLiteで開始**: 最速で機能実装に集中
2. **抽象化設計**: 後からPostgreSQL対応可能な設計
3. **必要時移行**: スケール要件が発生したら移行検討

### 具体的アクション
```bash
# 1. SQLiteで開始
go mod init chronome
go get gorm.io/driver/sqlite
go get gorm.io/gorm

# 2. 設定で切り替え可能な設計
# 3. 必要に応じてPostgreSQL対応追加
go get gorm.io/driver/postgres
```

この戦略により、**学習効率を最大化しつつ、将来の拡張性も確保**できます。

SQLiteから始めることで、データベース設定に時間を取られることなく、クリーンアーキテクチャの実装とビジネスロジックの開発に集中できるでしょう。

s0nmy: 現在の設計を厳守して以下の昨日の必要な部分を抜き出し
/docs に画面設計のmdファイルを作成してください


## 1. **ログイン画面**
- **要素:**
  - メールアドレス入力フィールド
  - パスワード入力フィールド
  - ログインボタン
  - 新規登録リンク
  - パスワードリセットリンク
- **機能:** 認証処理、セッション管理

## 2. **ダッシュボード画面**
- **要素:**
  - ヘッダー(ユーザー名、ログアウトボタン)
  - サイドバーナビゲーション
  - 今月の勤務統計サマリー(総労働時間、残業時間、出勤日数)
  - カレンダービュー
  - 最近の打刻履歴リスト
- **機能:** 勤怠状況の概要表示

## 3. **打刻画面**
- **要素:**
  - 現在時刻表示(リアルタイム更新)
  - 出勤ボタン
  - 退勤ボタン
  - 休憩開始/終了ボタン
  - 今日の打刻履歴表示
  - ステータス表示(勤務中/休憩中/退勤済み)
- **機能:** 出退勤・休憩時刻の記録

## 4. **勤怠履歴画面**
- **要素:**
  - 月選択ドロップダウン
  - カレンダービューまたはリストビュー切替
  - 日別勤怠記録テーブル
    - 日付
    - 出勤時刻
    - 退勤時刻
    - 休憩時間
    - 労働時間
    - ステータス
  - 編集・修正申請ボタン
  - CSV/PDFエクスポートボタン
- **機能:** 過去の勤怠データ閲覧、修正申請

## 5. **勤怠修正申請画面**
- **要素:**
  - 対象日選択
  - 修正項目選択(出勤/退勤/休憩)
  - 修正前時刻表示
  - 修正後時刻入力
  - 修正理由入力欄
  - 申請ボタン
  - キャンセルボタン
- **機能:** 打刻ミスの修正申請

## 6. **申請承認画面(管理者用)**
- **要素:**
  - 未承認申請リスト
  - 申請内容詳細表示
  - 承認ボタン
  - 却下ボタン
  - コメント入力欄
  - フィルター(申請者、日付、ステータス)
- **機能:** 部下の修正申請承認処理

## 7. **レポート画面**
- **要素:**
  - 期間選択(月次/年次)
  - グラフ表示
    - 月別労働時間推移
    - 残業時間推移
    - 部門別勤務状況
  - 集計テーブル
  - エクスポート機能
- **機能:** 勤怠データの統計・分析

## 8. **設定画面**
- **要素:**
  - プロフィール情報編集
  - 勤務時間設定(所定労働時間、休憩時間)
  - 通知設定
  - パスワード変更
  - 表示設定(テーマ、言語)
- **機能:** ユーザー設定のカスタマイズ

## 9. **管理画面(管理者用)**
- **要素:**
  - ユーザー管理テーブル
  - 部門管理
  - 勤務体系設定
  - システム設定
  - 監査ログ
- **機能:** システム全体の管理

## 共通UI要素
- **レスポンシブデザイン:** モバイル・タブレット・デスクトップ対応
- **ローディングインジケーター**
- **エラーメッセージ表示**
- **確認ダイアログ**
- **トースト通知**

s0nmy: これらのUIに酷似したWebアプリケーションを作成するのに適切なプロンプトを考えて

GitHub Copilot: # Webアプリケーション作成用プロンプト

以下は、開発者やAIアシスタントに指示する際の詳細なプロンプトです。

## 基本プロンプト

```
勤怠管理システムのWebアプリケーションを作成してください。

【技術スタック】
- フロントエンド: React + TypeScript + Tailwind CSS
- バックエンド: Node.js + Express
- データベース: PostgreSQL
- 認証: JWT
- 状態管理: React Context API または Redux

【主要機能】
1. ユーザー認証(ログイン/ログアウト)
2. 出退勤打刻機能(リアルタイム時計表示付き)
3. 休憩時間記録
4. 月次勤怠履歴閲覧
5. 勤怠修正申請機能
6. 管理者による承認機能
7. 勤怠レポート生成(CSV/PDFエクスポート)
8. ダッシュボード(統計サマリー表示)

【画面構成】
- ログイン画面
- ダッシュボード
- 打刻画面(出勤/退勤/休憩開始/終了ボタン)
- 勤怠履歴画面(カレンダービュー/リストビュー)
- 修正申請画面
- 承認画面(管理者用)
- レポート画面(グラフ・集計表)
- 設定画面
- 管理画面(ユーザー管理)

【UI/UX要件】
- レスポンシブデザイン(モバイルファースト)
- モダンで直感的なUI
- リアルタイム更新機能
- ローディング状態の表示
- エラーハンドリングとユーザーフィードバック
- トースト通知
- 確認ダイアログ

【データモデル】
- Users(ユーザー情報)
- TimeRecords(打刻記録)
- ModificationRequests(修正申請)
- Departments(部門)
- WorkSchedules(勤務体系)
```

## 段階的開発プロンプト

### Phase 1: プロジェクトセットアップ

```
勤怠管理アプリのプロジェクトを初期化してください。

1. Reactプロジェクトを作成(Vite + TypeScript)
2. Tailwind CSSを設定
3. 必要なパッケージをインストール:
   - react-router-dom(ルーティング)
   - axios(API通信)
   - date-fns(日付処理)
   - react-hook-form(フォーム管理)
   - recharts(グラフ表示)
   - zustand(状態管理)
4. フォルダ構成を作成:
   /src
     /components (共通コンポーネント)
     /pages (画面コンポーネント)
     /hooks (カスタムフック)
     /services (API通信)
     /types (型定義)
     /utils (ユーティリティ)
     /contexts (Context API)
```

### Phase 2: 認証機能

```
認証システムを実装してください。

【要件】
1. ログイン画面コンポーネント
   - メールアドレス・パスワード入力
   - バリデーション(react-hook-form使用)
   - エラー表示
2. JWT認証の実装
   - ローカルストレージにトークン保存
   - 自動ログイン機能
   - トークン期限切れハンドリング
3. ProtectedRoute コンポーネント
4. AuthContext で認証状態管理
5. ログアウト機能
```

### Phase 3: 打刻機能

```
打刻画面を実装してください。

【要件】
1. リアルタイム時計表示(1秒ごとに更新)
2. 出勤/退勤/休憩開始/休憩終了ボタン
3. 現在のステータス表示(勤務中/休憩中/退勤済み)
4. 本日の打刻履歴表示
5. ボタンの状態制御:
   - 出勤前は出勤ボタンのみ有効
   - 勤務中は休憩開始と退勤ボタン有効
   - 休憩中は休憩終了ボタンのみ有効
6. 打刻成功時のトースト通知
7. 位置情報取得(オプション)
```

### Phase 4: 勤怠履歴画面

```
勤怠履歴画面を実装してください。

【要件】
1. 月選択ドロップダウン
2. カレンダービューとリストビューの切り替え
3. 日別勤怠テーブル:
   - 日付、出勤時刻、退勤時刻、休憩時間、労働時間
   - 土日祝日のハイライト
   - 修正申請済みレコードのバッジ表示
4. 月次サマリー(総労働時間、残業時間、出勤日数)
5. CSVエクスポート機能
6. ページネーション
7. 修正申請ボタン(各レコード)
```

### Phase 5: ダッシュボード

```
ダッシュボード画面を実装してください。

【要件】
1. 今月の統計カード表示:
   - 総労働時間
   - 残業時間
   - 出勤日数
   - 平均勤務時間
2. 週間勤務時間グラフ(recharts使用)
3. 最近5件の打刻履歴リスト
4. カレンダー形式の出勤状況表示
5. クイックアクション(打刻ボタンへのリンク)
6. 未承認の修正申請件数バッジ
```

### Phase 6: 管理者機能

```
管理者用の画面を実装してください。

【要件】
1. 修正申請一覧画面:
   - 申請者、申請日、修正内容、ステータス
   - フィルター機能(未承認/承認済み/却下)
   - 詳細モーダル表示
2. 承認/却下機能:
   - 確認ダイアログ
   - コメント入力
3. ユーザー管理画面:
   - ユーザー一覧テーブル
   - 新規ユーザー登録
   - 編集/削除機能
   - 役割設定(一般/管理者)
4. 部門管理機能
```

### Phase 7: レポート機能

```
レポート画面を実装してください。

【要件】
1. 期間選択(月次/年次/カスタム期間)
2. グラフ表示:
   - 月別労働時間推移(折れ線グラフ)
   - 残業時間推移
   - 部門別勤務状況(棒グラフ)
3. 集計テーブル
4. PDF/CSVエクスポート機能
5. 印刷用スタイル
```

## バックエンドAPI プロンプト

```
勤怠管理システムのREST APIを実装してください。

【技術】
- Node.js + Express + TypeScript
- PostgreSQL + Prisma ORM
- JWT認証
- bcrypt(パスワードハッシュ化)

【エンドポイント】
POST   /api/auth/login
POST   /api/auth/register
POST   /api/auth/logout
GET    /api/auth/me

POST   /api/timecard/clock-in
POST   /api/timecard/clock-out
POST   /api/timecard/break-start
POST   /api/timecard/break-end
GET    /api/timecard/records?month=YYYY-MM
GET    /api/timecard/today

POST   /api/modification-requests
GET    /api/modification-requests
PATCH  /api/modification-requests/:id/approve
PATCH  /api/modification-requests/:id/reject

GET    /api/reports/monthly?month=YYYY-MM
GET    /api/reports/export/csv
GET    /api/reports/export/pdf

GET    /api/admin/users
POST   /api/admin/users
PATCH  /api/admin/users/:id
DELETE /api/admin/users/:id

【ミドルウェア】
- 認証チェック
- 管理者権限チェック
- エラーハンドリング
- リクエストロギング
- レート制限
```

## データベーススキーマプロンプト

```
勤怠管理システムのデータベーススキーマを設計してください。

【テーブル】

users:
- id (UUID)
- email (unique)
- password_hash
- name
- role (enum: 'user', 'admin')
- department_id (FK)
- created_at
- updated_at

time_records:
- id (UUID)
- user_id (FK)
- date (DATE)
- clock_in (TIMESTAMP)
- clock_out (TIMESTAMP)
- break_start (TIMESTAMP)
- break_end (TIMESTAMP)
- total_break_minutes (INT)
- total_work_minutes (INT)
- status (enum: 'working', 'break', 'completed')
- created_at
- updated_at

modification_requests:
- id (UUID)
- user_id (FK)
- time_record_id (FK)
- field_name (VARCHAR)
- old_value (TIMESTAMP)
- new_value (TIMESTAMP)
- reason (TEXT)
- status (enum: 'pending', 'approved', 'rejected')
- approved_by (FK to users)
- approved_at (TIMESTAMP)
- comment (TEXT)
- created_at

departments:
- id (UUID)
- name (VARCHAR)
- created_at

【インデックス】
- users.email
- time_records.user_id, date
- modification_requests.user_id, status
```


GitHub Copilot: # 画面設計書

最終更新: 2025年11月10日  
対象: ChronoMe (一人用タイムカード Web アプリケーション)

---

## 概要

本ドキュメントは ChronoMe の画面設計を定義します。本アプリは**個人利用**を前提とし、複数ユーザー管理や承認フロー等のチーム機能は含みません。

### 前提条件
- **ユーザー数**: 1名（個人利用）
- **認証方式**: サインド Cookie ベースのセッション
- **技術スタック**: React + TypeScript + Tailwind CSS
- **レスポンシブ対応**: モバイル・タブレット・デスクトップ

---

## 画面一覧

| 画面名 | パス | 認証 | 説明 |
|--------|------|------|------|
| ログイン | `/login` | 不要 | ユーザー認証 |
| サインアップ | `/signup` | 不要 | 新規ユーザー登録 |
| ダッシュボード | `/` | 必須 | 作業時間の概要・クイック操作 |
| タイマー | `/timer` | 必須 | 作業開始/終了・リアルタイム計測 |
| 履歴 | `/history` | 必須 | 過去の作業記録一覧・編集 |
| レポート | `/reports` | 必須 | 日次・週次・月次の集計 |
| プロジェクト管理 | `/projects` | 必須 | プロジェクトCRUD |
| タグ管理 | `/tags` | 必須 | タグCRUD |
| 設定 | `/settings` | 必須 | プロフィール・タイムゾーン設定 |

---

## 1. ログイン画面

### パス
`/login`

### 目的
既存ユーザーの認証処理

### UI要素
```tsx
- ロゴ・アプリ名表示
- メールアドレス入力フィールド
  - type: email
  - required
  - バリデーション: email形式
- パスワード入力フィールド
  - type: password
  - required
  - 表示/非表示切り替えアイコン
- ログインボタン
  - ローディング状態表示
- 新規登録リンク
  - → `/signup` へ遷移
- エラーメッセージ表示エリア
```

### 機能要件
1. フォームバリデーション（react-hook-form使用）
2. POST `/api/auth/login` でサインド Cookie 取得
3. 認証成功時にダッシュボードへリダイレクト
4. 認証失敗時にエラーメッセージ表示（「メールアドレスまたはパスワードが正しくありません」）

### レスポンシブ対応
- モバイル: 全幅カード、垂直レイアウト
- タブレット以上: 中央配置カード（max-width: 400px）

---

## 2. サインアップ画面

### パス
`/signup`

### 目的
新規ユーザー登録

### UI要素
```tsx
- ロゴ・アプリ名表示
- メールアドレス入力
  - バリデーション: email形式、重複チェック
- パスワード入力
  - バリデーション: 最小8文字、英数字混在
  - 強度インジケーター表示
- パスワード確認入力
  - バリデーション: 一致チェック
- 表示名入力（任意）
  - 最大50文字
- タイムゾーン選択
  - デフォルト: ブラウザ検出値
  - IANA timezone一覧から選択
- 登録ボタン
- ログインリンク
  - → `/login` へ遷移
```

### 機能要件
1. POST `/api/auth/signup` で新規登録
2. 登録成功時に自動ログイン → ダッシュボードへ
3. バリデーションエラーは各フィールド下に表示
4. メールアドレス重複時は「このメールアドレスは既に登録されています」

---

## 3. ダッシュボード画面

### パス
`/` (認証必須)

### 目的
作業状況の概要表示とクイック操作

### レイアウト
```
┌─────────────────────────────────────┐
│ ヘッダー: ユーザー名・設定・ログアウト │
├─────────────────────────────────────┤
│ サイドバー │ メインコンテンツ         │
│ - ダッシュ │ ┌─────────────────┐   │
│ - タイマー │ │ 今月の統計カード │   │
│ - 履歴     │ │ - 総労働時間     │   │
│ - レポート │ │ - 作業日数       │   │
│ - プロジェ │ │ - 平均勤務時間   │   │
│ - タグ     │ └─────────────────┘   │
│ - 設定     │                         │
│            │ ┌─────────────────┐   │
│            │ │ 今週の作業時間   │   │
│            │ │ (折れ線グラフ)   │   │
│            │ └─────────────────┘   │
│            │                         │
│            │ ┌─────────────────┐   │
│            │ │ 最近の作業記録   │   │
│            │ │ (リスト5件)      │   │
│            │ └─────────────────┘   │
│            │                         │
│            │ [作業を開始する]        │
└─────────────────────────────────────┘
```

### UI要素

#### 統計カード
```tsx
- 今月の総労働時間
  - HH:MM:SS 形式
  - 前月比較バッジ（±X時間）
- 今月の作業日数
  - X日 / 営業日数
- 今週の平均勤務時間
  - HH:MM形式
```

#### 今週の作業時間グラフ
```tsx
- 横軸: 月曜〜日曜
- 縦軸: 時間（時）
- recharts の BarChart 使用
- 休憩時間は除外
- is_break=true のエントリは集計から除外
```

#### 最近の作業記録リスト
```tsx
各レコード表示項目:
- プロジェクト名（color でバッジ表示）
- 作業タイトル
- 開始時刻 - 終了時刻
- 作業時間（duration_sec から計算）
- タグ（複数可、color でバッジ）

アクション:
- クリックで詳細モーダル表示
- 編集アイコン → 履歴画面へ遷移
```

#### クイックアクション
```tsx
- [作業を開始する] ボタン
  - → `/timer` へ遷移
  - 未終了エントリがある場合は確認ダイアログ表示
```

### レスポンシブ対応
- モバイル: サイドバーはハンバーガーメニュー、カード縦並び
- タブレット以上: サイドバー固定表示、カード横並び

---

## 4. タイマー画面

### パス
`/timer` (認証必須)

### 目的
作業時間のリアルタイム計測と記録

### UI要素

#### リアルタイム時計
```tsx
- 現在時刻表示（1秒ごと更新）
- フォーマット: YYYY年MM月DD日 HH:MM:SS
- ユーザーの time_zone 設定に基づいて表示
```

#### 作業状態表示
```tsx
- ステータスバッジ
  - 未開始: グレー
  - 作業中: グリーン、経過時間表示
  - 休憩中: オレンジ、休憩時間表示
  - 終了済み: ブルー
```

#### 作業開始フォーム
```tsx
表示条件: 未開始状態のみ

- 作業タイトル入力
  - required
  - 最大100文字
- プロジェクト選択
  - ドロップダウン
  - required
  - 新規作成リンク
- タグ選択（複数可）
  - マルチセレクト
  - 新規作成可能
- [作業を開始する] ボタン
  - POST `/api/entries/start`
  - started_at に現在時刻を記録
```

#### 作業中の操作
```tsx
表示条件: 作業中または休憩中

- 経過時間表示（MM:SS形式、1秒更新）
- [休憩開始] / [休憩終了] ボタン
  - 休憩開始: 新しい is_break=true のエントリを作成
  - 休憩終了: 休憩エントリの ended_at を記録
- [作業を終了する] ボタン
  - POST `/api/entries/{id}/stop`
  - ended_at に現在時刻を記録
  - duration_sec を自動計算
```

#### 本日の作業履歴
```tsx
- 今日開始された全エントリを時系列表示
- 作業エントリと休憩エントリを区別
  - 作業: 青いバー
  - 休憩: グレーのバー
- 各エントリ表示項目:
  - 開始時刻 - 終了時刻（未終了は「進行中」）
  - 作業時間 / 休憩時間
  - プロジェクト・タグ
- クイック編集機能（時刻修正・削除）
```

### 並行作業対応
```tsx
- 複数の作業を同時開始可能
- ratio フィールドで時間配分を後から調整
- UIでは後ほどガントチャート形式で調整機能を提供予定
```

### 機能要件
1. リアルタイム時計は `useEffect` + `setInterval` で実装
2. 経過時間は started_at からの差分を1秒ごと計算
3. 作業開始時に未終了エントリがある場合の挙動:
   - 並行作業として許容
   - 確認ダイアログなし（将来的に ratio 調整機能で対応）
4. 休憩時間は is_break=true で区別
5. トースト通知で操作完了をフィードバック

---

## 5. 履歴画面

### パス
`/history` (認証必須)

### 目的
過去の作業記録の閲覧・編集

### UI要素

#### フィルター・ソート
```tsx
- 期間選択
  - 今週 / 今月 / 先月 / カスタム期間
  - カスタム: date picker で from/to 指定
- プロジェクトフィルター
  - マルチセレクト
  - "全て" オプション
- タグフィルター
  - マルチセレクト
- 表示形式切り替え
  - カレンダービュー / リストビュー
- ソート
  - 新しい順 / 古い順 / 作業時間順
```

#### リストビュー
```tsx
テーブル列:
- 日付
- 開始時刻
- 終了時刻
- 作業時間（duration_sec から算出、休憩除外）
- プロジェクト
- タイトル
- タグ
- 操作（編集・削除アイコン）

機能:
- ページネーション（20件/ページ）
- 行クリックで詳細モーダル
- 編集アイコンで編集モーダル
- 削除アイコンで確認ダイアログ
```

#### カレンダービュー
```tsx
- react-big-calendar 等を使用
- 各日のセルに作業記録を表示
  - プロジェクトカラーでバー表示
  - 複数作業がある場合は積み重ね
- 日付クリックでその日の詳細モーダル
- 土日は背景色変更
```

#### 編集モーダル
```tsx
編集可能項目:
- 作業タイトル
- 開始時刻
- 終了時刻
- プロジェクト
- タグ
- メモ
- 休憩フラグ (is_break)

バリデーション:
- 終了時刻 > 開始時刻
- 必須項目チェック

操作:
- [保存] ボタン: PATCH `/api/entries/{id}`
- [キャンセル] ボタン
- [削除] ボタン: 確認ダイアログ → DELETE `/api/entries/{id}`
```

#### 月次サマリー
```tsx
選択期間の集計表示:
- 総作業時間
- 作業日数
- プロジェクト別内訳（円グラフ）
- タグ別内訳（棒グラフ）

エクスポート:
- [CSVダウンロード] ボタン
  - GET `/api/reports/export?format=csv&from=...&to=...`
```

### レスポンシブ対応
- モバイル: リストビューのみ、テーブルをカード形式で表示
- タブレット以上: カレンダービュー利用可能

---

## 6. レポート画面

### パス
`/reports` (認証必須)

### 目的
作業時間の統計・分析

### UI要素

#### タブ切り替え
```tsx
- 日次レポート
- 週次レポート
- 月次レポート
```

#### 日次レポート
```tsx
- 日付選択（date picker）
- 当日の作業一覧
  - タイムライン形式表示
  - 作業と休憩を区別（色分け）
- 統計カード:
  - 総作業時間
  - 休憩時間
  - 実働時間
  - プロジェクト別内訳
```

#### 週次レポート
```tsx
- 週選択（前週/今週/次週ナビゲーション）
- 曜日別作業時間グラフ（棒グラフ）
- プロジェクト別内訳（円グラフ）
- 日別作業時間テーブル
- 週間サマリー:
  - 総作業時間
  - 平均勤務時間
  - 最長/最短勤務日
```

#### 月次レポート
```tsx
- 月選択（ドロップダウン）
- 日別作業時間推移（折れ線グラフ）
- プロジェクト別累計時間（棒グラフ）
- タグ別分析
- 月間サマリー:
  - 総作業時間
  - 作業日数
  - 平均勤務時間
  - 最も使用したプロジェクト/タグ
```

#### エクスポート機能
```tsx
- [CSVダウンロード] ボタン
  - 選択期間のデータをCSV形式で出力
  - ヘッダー: 日付, 開始時刻, 終了時刻, 作業時間, プロジェクト, タイトル, タグ
- [印刷] ボタン
  - 印刷用CSS適用
  - グラフを含む
```

### 機能要件
1. グラフは recharts を使用
2. 休憩時間 (is_break=true) は集計から除外
3. 期間選択変更時に自動再読み込み
4. データがない期間は「データがありません」表示
5. ローディング状態を適切に表示

---

## 7. プロジェクト管理画面

### パス
`/projects` (認証必須)

### 目的
プロジェクトのCRUD操作

### UI要素

#### プロジェクト一覧
```tsx
表示形式: カード形式またはテーブル

各プロジェクトカード:
- プロジェクト名
- カラーバッジ
- アーカイブ状態
- 統計:
  - 累計作業時間
  - 直近の作業日
  - 使用回数
- 操作アイコン:
  - 編集
  - アーカイブ/アーカイブ解除
  - 削除（確認ダイアログ）

フィルター:
- [アーカイブ済みを表示] チェックボックス
```

#### 新規作成ボタン
```tsx
- [+ 新しいプロジェクト] ボタン
- クリックで作成モーダル表示
```

#### 作成/編集モーダル
```tsx
フォーム項目:
- プロジェクト名
  - required
  - 最大80文字
  - ユーザー内ユニーク
- カラー選択
  - カラーピッカー
  - デフォルト: #3B82F6
  - HEX形式 (#RRGGBB)
- アーカイブ状態（編集時のみ）
  - チェックボックス

操作:
- [保存] ボタン
  - 新規: POST `/api/projects`
  - 編集: PATCH `/api/projects/{id}`
- [キャンセル] ボタン

バリデーション:
- 名前重複チェック
- カラー形式チェック
```

### 機能要件
1. プロジェクト削除時に紐付くエントリの確認
   - エントリが存在する場合は警告ダイアログ
   - 「削除すると関連する作業記録からプロジェクト情報が失われます」
2. アーカイブ済みプロジェクトは一覧で既定非表示
3. カラーピッカーは react-colorful 等を使用

---

## 8. タグ管理画面

### パス
`/tags` (認証必須)

### 目的
タグのCRUD操作

### UI要素
プロジェクト管理画面と同様の構成

```tsx
各タグカード:
- タグ名
- カラーバッジ
- 統計:
  - 使用回数
  - 直近の使用日
- 操作アイコン:
  - 編集
  - 削除

作成/編集モーダル:
- タグ名（required, 最大30文字, ユーザー内ユニーク）
- カラー選択（HEX形式）
```

### 機能要件
1. タグ削除時は entry_tags の関連レコードも削除
2. 削除確認ダイアログで使用中のエントリ数を表示

---

## 9. 設定画面

### パス
`/settings` (認証必須)

### 目的
ユーザー設定のカスタマイズ

### UI要素

#### プロフィール設定
```tsx
- 表示名
  - 最大50文字
- メールアドレス
  - email形式
  - 変更時は再認証が必要（将来実装）
- タイムゾーン選択
  - IANA timezone
  - 変更時は確認ダイアログ（「時刻表示が変更されます」）
- [保存] ボタン
  - PATCH `/api/users/me`
```

#### パスワード変更
```tsx
- 現在のパスワード
  - required
- 新しいパスワード
  - required
  - 最小8文字、英数字混在
  - 強度インジケーター
- 新しいパスワード（確認）
  - 一致チェック
- [パスワードを変更] ボタン
  - POST `/api/users/me/password`
  - 成功時は自動ログアウト
```

#### 表示設定（将来実装）
```tsx
- テーマ選択（ライト/ダーク/システム）
- 言語選択（日本語/English）
```

#### データ管理
```tsx
- [全データをエクスポート] ボタン
  - JSON形式でダウンロード
- [アカウントを削除] ボタン
  - 確認ダイアログ（警告表示）
  - DELETE `/api/users/me`
  - 削除後はログイン画面へリダイレクト
```

---

## 共通UI要素

### ヘッダー
```tsx
全画面共通:
- アプリロゴ/名称
- ユーザー名表示
- 設定アイコン → `/settings`
- ログアウトボタン
  - POST `/api/auth/logout`
  - Cookie削除
  - ログイン画面へリダイレクト
```

### サイドバーナビゲーション
```tsx
メニュー項目:
- ダッシュボード
- タイマー
- 履歴
- レポート
- プロジェクト
- タグ
- 設定

モバイル:
- ハンバーガーメニュー
- オーバーレイで表示
- 外側クリックで閉じる

デスクトップ:
- 固定表示
- 折りたたみ可能
```

### ローディングインジケーター
```tsx
使用場面:
- API通信中
- データ読み込み中
- ページ遷移中

種類:
- ページ全体: スピナー + 「読み込み中...」
- ボタン内: 小スピナー + ボタン無効化
- 部分的: スケルトンスクリーン
```

### エラー表示
```tsx
パターン:
1. フォームバリデーションエラー
   - フィールド下に赤文字で表示
2. API エラー
   - トースト通知（3秒間表示）
   - エラーメッセージ + リトライボタン
3. ページエラー
   - エラーバウンダリで捕捉
   - エラーページ表示 + ホームに戻るボタン
```

### トースト通知
```tsx
使用場面:
- 作業開始/終了成功
- データ保存成功
- エラー発生
- 削除成功

位置: 画面右上
自動消失: 3秒
種類:
- success: 緑
- error: 赤
- info: 青
- warning: 黄
```

### 確認ダイアログ
```tsx
使用場面:
- 削除操作
- アーカイブ/アーカイブ解除
- アカウント削除

構成:
- タイトル
- 説明文
- [キャンセル] ボタン
- [実行] ボタン（赤色、危険な操作）
```

### モーダル
```tsx
共通仕様:
- オーバーレイ背景（半透明黒）
- 中央配置
- ESCキーで閉じる
- 外側クリックで閉じる
- ヘッダー（タイトル + 閉じるボタン）
- コンテンツエリア
- フッター（アクションボタン）
```

---

## レスポンシブブレークポイント

Tailwind CSS のデフォルトに準拠:

```tsx
sm: 640px   // スマートフォン横向き
md: 768px   // タブレット縦向き
lg: 1024px  // タブレット横向き
xl: 1280px  // デスクトップ
2xl: 1536px // 大型デスクトップ
```

### 各画面の対応
- **モバイル（< md）**: カード縦並び、サイドバーハンバーガーメニュー
- **タブレット（md ～ lg）**: 2カラムレイアウト、サイドバー表示
- **デスクトップ（lg ～）**: 3カラムレイアウト可能、サイドバー固定

---

## アクセシビリティ要件

1. **キーボード操作**
   - Tab キーでフォーカス移動
   - Enter/Space でボタン実行
   - ESC でモーダル/ドロップダウンを閉じる

2. **ARIA属性**
   - ボタン: `aria-label` で目的を明示
   - フォーム: `aria-describedby` でエラー関連付け
   - モーダル: `role="dialog"`, `aria-modal="true"`

3. **コントラスト比**
   - WCAG AA基準（4.5:1以上）

4. **フォーカスインジケーター**
   - フォーカス時に視覚的な枠線表示

---

## パフォーマンス要件

1. **初期表示**: 3秒以内
2. **画面遷移**: 1秒以内
3. **API レスポンス**: 500ms以内（目標）
4. **リストページ**: 仮想スクロール（react-window）を検討（将来）

---

## セキュリティ考慮事項

1. **XSS対策**
   - ユーザー入力は React の自動エスケープに依存
   - dangerouslySetInnerHTML は使用しない

2. **CSRF対策**
   - SameSite=Lax Cookie
   - 状態変更は POST/PATCH/DELETE のみ

3. **認証チェック**
   - ProtectedRoute コンポーネントで全画面保護
   - 未認証時はログイン画面へリダイレクト

---

## 今後の拡張（参考）

本ドキュメントの範囲外ですが、以下の機能を検討:

- PWA対応（オフライン機能）
- プッシュ通知
- 音声入力
- Slack連携
- チーム機能（複数ユーザー、承認フロー）

---

## まとめ

ChronoMe の画面設計は、個人利用を前提としたシンプルで直感的な UI を目指します。クリーンアーキテクチャと整合し、API 設計（APIDesign.md）およびデータベース設計（DBDesign.md）に基づいて実装してください。