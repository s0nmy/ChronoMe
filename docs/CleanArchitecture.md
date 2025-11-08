# ChronoMe クリーンアーキテクチャ実装ガイド

本ドキュメントは `docs/DesignDoc.md` から切り出したクリーンアーキテクチャの詳細実装方法です。ChronoMe の Go バックエンド実装における具体的なコード例、ベストプラクティス、テスト戦略を提供します。

---

## 1. アーキテクチャ概観と実装方針

```mermaid
flowchart LR
    UI[HTTP Handler<br>(Chi)] --> UC[Usecase]
    UC -->|依存| ENT[Entity]
    UC -->|ポート| REPO_IF[Repository Interface]
    REPO_IMPL[Repository (GORM)] --> REPO_IF
    REPO_IMPL --> DB[(SQLite / PostgreSQL)]
    UC --> PRES[Response DTO]
    PRES --> UI
```

### 実装における重要な原則

1. **Interface設計の簡素化**: 過度に複雑なInterface分割を避け、必要最小限のメソッドに留める
2. **Mock最小化**: 可能な限りシンプルなスタブやインメモリ実装で検証し、必要に応じて SQLite での軽量統合テスト（将来的には PostgreSQL でも同パッケージを利用）を行う
3. **時刻依存対応の簡素化**: `TimeProvider` インターフェースは `Now()` のみに絞り、実装とテストで容易に差し替えられる構成にする

- 依存性は常に内向き（UI → Usecase → Entity）となるように保つ。
- 外部リソース（DB や外部 API）は Infrastructure 層に隔離し、Usecase 層からはポート (interface) を介してアクセスする。

---

## 2. 層ごとの責務

| 層 | ディレクトリ例 | 主な責務 | 禁止事項 |
|----|----------------|----------|----------|
| Entity | `internal/entity` | ドメインモデルとビジネスルールを表現。値オブジェクトやドメインサービスを含む。 | フレームワーク依存、入出力形式への依存 |
| Usecase | `internal/usecase` | アプリケーション固有のユースケースを実装し、トランザクション制御やバリデーション、リトライポリシーを管理。 | ORM や HTTP の直接利用、DB スキーマへの直接依存 |
| Interface Adapter | `internal/handler`, `internal/presenter`, `internal/repository` | 入出力変換（DTO ↔ Entity）、REST ハンドラ、永続化のインターフェース実装。 | Usecase 層を飛び越えて Infrastructure に直接依存すること |
| Infrastructure | `pkg/infra`, `internal/repository/gorm`, `cmd/server` | フレームワーク設定、DB クライアント構築、依存の組み立て。 | 内側の層を知らない型を勝手に生成しない |

---

## 3. 依存性ルール

1. **内向き依存のみ**  
   - `usecase` は `entity` へ依存可、逆は禁止。  
   - `handler` / `repository` は `usecase` インターフェースへ依存。  
   - インターフェースによる DIP (Dependency Inversion Principle) を徹底する。

2. **DTO と Entity の分離**  
   - REST で受け取る JSON は `handler` 層で DTO にマッピングし、`usecase` へ受け渡す。  
   - `usecase` から UI へ返却する際は Presenter を通じてレスポンス整形する。

3. **トランザクション境界**  
   - トランザクション開始／終了は `usecase` 層で interface (`TransactionManager`) を通して制御する。  
   - Repository 層は受け取った `context.Context` 内の `Tx` に基づき処理を行う。

4. **時刻・設定の注入**  
   - 現在時刻や設定値は `usecase` に `TimeProvider` / `Config` インターフェースを注入し、副作用をテストで差し替え可能にする。`TimeProvider` は単一メソッド (`Now() time.Time`) を持つ最小限の契約とする。

---

## 4. データフロー

1. **入力**: `handler` が HTTP リクエストを受信 → DTO バリデーション → `usecase` 呼び出し。  
2. **アプリケーションロジック**: `usecase` がドメインロジックを実行し、必要な Repository interface を利用。  
3. **永続化**: Repository 実装が GORM を通じて SQLite（初期）へアクセスし、ドライバー切替で PostgreSQL にも対応。  
4. **出力**: `usecase` が結果を返却 → Presenter がレスポンス DTO へ変換 → `handler` から HTTP 応答。

Request/Response DTO は `internal/handler/dto` などに分離し、`json` タグを持つ構造体を定義する。

---

## 5. コーディングガイドライン

- **ユースケースメソッド**は `context.Context` を第 1 引数に、ID や入力 DTO を続ける。  
- **Entity** はビジネスルールをメソッドで表現し、副作用（DBアクセス・ログ）は持たない。  
- **Repository Interface** はユースケースで必要な操作単位で定義し、低レベル API へ引きずられない。  
- **エラーハンドリング**  
  - ドメインエラーは `entity` で型化し、`usecase` で HTTP ステータスへマップ。  
  - インフラエラー（DB障害など）は `usecase` でラップして呼び出し元へ返す。  
- **ログ** は `usecase` または `handler` が行い、`entity` では行わない。  
- **コンフィグ読込** は `cmd/` または `pkg/infra` で行い、`usecase` にはインターフェース経由で渡す。


## 6. テスト戦略

| 層 | テスト種別 | 方針 |
|----|------------|------|
| Entity | 単体テスト | ドメインロジックの純粋なテスト。外部依存なし。 |
| Usecase | モックベース単体テスト | Repository をモック化し、副作用のないロジックを検証。 |
| Repository | 軽量統合テスト | SQLite（インメモリ/ファイル）で CRUD と制約を確認。 |
| Handler | API テスト | `httptest` で HTTP レスポンスとバリデーションを検証。 |

- テストダブルには `testify/mock` などを利用するが、`entity` 層では標準ライブラリのみを使用する。  
- `TimeProvider` などのインターフェースはテストで差し替え、固定時刻で検証できるようにする。

---

## 7. ディレクトリマッピング（例）

```
backend/
└── internal/
    ├── entity/
    │   └── user.go
    ├── usecase/
    │   ├── user_usecase.go
    │   └── entry_usecase.go
    ├── repository/
    │   ├── interface.go        // Port 定義
    │   └── gorm/               // GORM 実装
    │       ├── entry_repository.go
    │       ├── project_repository.go
    │       ├── user_repository.go
    │       └── models/
    │           ├── entry.go
    │           ├── project.go
    │           └── user.go
    ├── handler/
    │   ├── entry_handler.go
    │   └── dto/
    │       └── entry_request.go
    └── presenter/
        └── entry_presenter.go
```

- `cmd/server/main.go` で必要な依存を手動で初期化し、各コンストラクタへ渡す。  

依存性注入は自動生成ツールや DI コンテナを用いず、構造体のコンストラクタに引数として渡す方針です。これにより依存関係が明示的になり、テスト時にはモックやスタブを容易に差し替えられます。
- `pkg/infra/db` などに DB 接続や設定ローダを配置する想定。

---

## 8. トランザクション設計

- Usecase で `TransactionManager` インターフェースを呼び出し、複数 Repository 操作をひとまとめにする。  
- `TransactionManager.Execute(ctx, func(ctx context.Context) error)` のようなパターンを採用し、内部で `context` にトランザクションを格納する。  
- Repository 実装は `context` から `*gorm.DB`（または `*sql.Tx`）を取得し、一貫したトランザクションを使用する。

---

## 9. 変更管理

- 層の依存規則を破る変更（例: Handler から Repository 実装を直接呼ぶ）は PR レビューで検出する。  
- 新しいユースケース追加時は、まず Entity / Usecase に最低限のインターフェースを追加し、外部 I/O は後から組み込む。  
- 依存グラフを自動チェックする場合は `go list -deps` を活用し、`lint` スクリプトで禁止パッケージ参照を検出する。

---

## 10. 参考

- Robert C. Martin, *Clean Architecture*  
- `docs/DesignDoc.md`：全体技術設計  
- `docs/APIDesign.md`：REST API 詳細  
- `docs/DBDesign.md`：データベース設計

本ドキュメントはクリーンアーキテクチャの適用状況を定期レビューする際の参照資料として利用し、必要に応じて更新してください。

最終更新: 2025-10-22
