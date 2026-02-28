# Contributing

## Code of Conduct

このリポジトリは学習プロジェクトです。相互に敬意を払い、建設的なコミュニケーションを心がけてください。

## PR 作成フロー

1. Issue/タスクを確認して作業内容を明確にします。
2. `main` から作業ブランチを作成します。
3. 変更をコミットし、PR を作成します。
4. 可能な限りテストと lint を実行します。

### コミットメッセージ

コミットメッセージの詳細ルールは [docs/CommitGuidelines.md](docs/CommitGuidelines.md) を参照してください。

## コーディングスタンダード

- Go: `gofmt` を適用し、パッケージ設計は Clean Architecture を維持します。
- TypeScript/React: ESLint + Prettier の設定に従います。
- 変更は小さく、目的がわかる粒度でコミットします。

## ビルド手順

```bash
cd backend
go build ./cmd/server
```

```bash
cd frontend
npm install
npm run build
```

## テスト・静的解析

```bash
cd backend
go test ./...
```

```bash
cd frontend
npm run lint
```

フロントエンドには現時点で自動テストがないため、変更時は lint の実行を推奨します。

## 環境変数

### バックエンド (Go)

| 変数 | 説明 | デフォルト |
| --- | --- | --- |
| `APP_ENV` | 実行環境 | `development` |
| `SERVER_ADDRESS` | バインドアドレス | `:8080` |
| `DB_DRIVER` | DB ドライバ | `sqlite` |
| `DB_DSN` | DB 接続先 | `dev.db` |
| `ALLOWED_ORIGIN` | CORS 許可 Origin | `http://localhost:3000` |
| `SESSION_SECRET` | セッション署名用シークレット | `dev-secret-change-me` |
| `SESSION_TTL` | セッション有効期限 | `12h` |
| `SESSION_COOKIE_SECURE` | Secure Cookie 有効化 | `false` (development) |
| `DEFAULT_PROJECT_COLOR` | プロジェクト初期色 | `#3B82F6` |

### フロントエンド (Vite)

| 変数 | 説明 | デフォルト |
| --- | --- | --- |
| `VITE_API_BASE_URL` | API ベース URL | `http://localhost:8080` |

## コード設計

- Go バックエンドは Handler / Usecase / Repository の層構造を維持します。
- React フロントエンドは `features/` 単位で機能を分割し、UI は `components/` に集約します。
- 仕様や設計の詳細は `.docs/DesignDoc.md` を参照してください。
