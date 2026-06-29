# ChronoMe

ChronoMe は個人の作業時間を記録・集計するタイムカード Web アプリケーションです。
Go 製バックエンドと React + TypeScript フロントエンドで構成し、クリーンアーキテクチャを採用しています。

## 概要

- フロントエンド: React + TypeScript (Vite)
- バックエンド: Go
- Allocation API: Go バックエンド内の `/api/allocations`
- iOS: SwiftUI（iOS 17以上）

## 使い方

ローカル開発では以下を実行します。

```bash
make dev
```

```bash
make backend  # or make b
make frontend # or make f
```

### iOS

[`ios/ChronoMe.xcodeproj`](ios/ChronoMe.xcodeproj) を Xcode で開き、`ChronoMe`
scheme と iPhone Simulator を選択して実行します。初回は Swift Package Manager
による依存関係の解決が行われます。

## ドキュメント

- PRD: [`docs/PRD.md`](docs/PRD.md)
- 設計ドキュメント: [`docs/DesignDoc.md`](docs/DesignDoc.md)
- API/仕様: [`docs/`](docs/)
- コミットルール: [`docs/CommitGuidelines.md`](docs/CommitGuidelines.md)

## Contributing

詳細は [CONTRIBUTING.md](CONTRIBUTING.md) と [`docs/DesignDoc.md`](docs/DesignDoc.md) を参照してください。

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
