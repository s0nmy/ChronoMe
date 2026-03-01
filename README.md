# ChronoMe

ChronoMe は個人の作業時間を記録・集計するタイムカード Web アプリケーションです。
Go 製バックエンドと React + TypeScript フロントエンドで構成し、クリーンアーキテクチャを採用しています。

## 概要

- フロントエンド: React + TypeScript (Vite)
- バックエンド: Go
- Allocation API: Go バックエンド内の `/api/allocations`

## 使い方

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

## ドキュメント

- 設計ドキュメント: [`.docs/DesignDoc.md`](.docs/DesignDoc.md)
- API/仕様: [`docs/`](docs/)
- コミットルール: [`docs/CommitGuidelines.md`](docs/CommitGuidelines.md)

## Contributing

詳細は [CONTRIBUTING.md](CONTRIBUTING.md) と [`.docs/DesignDoc.md`](.docs/DesignDoc.md) を参照してください。

- PR / MR 手順: ブランチ作成 → テスト/静的解析 → PR/MR 作成
- コーディングスタンダード: Go は `gofmt`、フロントは ESLint と Prettier
- テスト:
  ```bash
  cd backend && go test ./...
  ```
- lint / 静的解析:
  ```bash
  cd frontend && npm run lint
  ```
