# ChronoMe

ChronoMe は個人の作業時間を記録・集計するタイムカード Web アプリケーションです。
Go 製バックエンドと React + TypeScript フロントエンドで構成し、クリーンアーキテクチャを採用しています。

## Usage

ローカル開発では以下のコマンドで各サービスを起動します。

```bash
cd backend
go run ./cmd/server
```

```bash
cd frontend
npm install
npm run dev
```

```bash
cd node-backend
npm install
npm run dev
```

### ローカルアクセス例

- フロントエンド: http://localhost:3000
- Go API: http://localhost:8080
- Allocation API: http://localhost:4000

詳細な仕様・設計ドキュメントは `.docs/DesignDoc.md` と `docs/` 配下の各ドキュメントを参照してください。

## Contributing

Refer to the documentation in [CONTRIBUING.md](CONTRIBUTING.md).

- PR / MR 手順: ブランチ作成 → テスト/静的解析 → PR/MR 作成
- ディレクトリ構成: `backend/`, `frontend/`, `node-backend/`, `docs/`, `.docs/`
- コーディングスタンダード: Go は `gofmt`、フロント/Node は ESLint と Prettier
- テスト: `cd backend && go test ./...`
- lint / 静的解析: `cd frontend && npm run lint`, `cd node-backend && npm run lint`
