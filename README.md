# ChronoMe

ChronoMe は個人の作業時間を記録・集計するタイムカード Web アプリケーションです。
Go 製バックエンドと React + TypeScript フロントエンドで構成し、クリーンアーキテクチャを採用しています。

## 概要

- フロントエンド: React + TypeScript (Vite)
- バックエンド: Go
- Allocation API: Go バックエンド内の `/api/allocations`
- iOS: SwiftUI（iOS 17以上）

## 使い方

### Supabase 認証の設定

認証済み API を利用するには Supabase プロジェクトが必要です。`.env.example` を
`.env` としてコピーし、プロジェクト URL、anon key、JWT signing secret を設定します。

```bash
cp .env.example .env
# .env 内の your-* を Supabase Dashboard の値に置き換える
set -a; source .env; set +a
```

`VITE_SUPABASE_URL` と `VITE_SUPABASE_ANON_KEY` はフロントエンドの起動時に必須です。
`SUPABASE_JWT_SECRET` はバックエンドが Bearer トークンを検証するために必須です。値が
未設定の場合、バックエンドは起動を停止します。

Supabase Dashboard では、開発用の Site URL を `http://localhost:3000`、Redirect URL を
`http://localhost:3000/auth/callback` に設定してください。メール確認を有効にするかどうかは
運用方針として決定し、OAuth プロバイダーごとの Redirect URL も同じ画面で登録します。

既存の Cookie 認証アカウントは、メールアドレスだけでは Supabase アカウントへ自動移行
されません。アカウント乗っ取りを防ぐため、明示的な連携・復旧フローを提供するまで既存
アカウントへのログインは拒否されます。

### ローカル開発

環境変数を読み込んだ状態で以下を実行します。

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

- PRD: [`docs/product/PRD.md`](docs/product/PRD.md)
- 設計ドキュメント: [`docs/product/DesignDoc.md`](docs/product/DesignDoc.md)
- API/仕様: [`docs/`](docs/)
- コミットルール: [`docs/development/CommitGuidelines.md`](docs/development/CommitGuidelines.md)

## Contributing

詳細は [CONTRIBUTING.md](CONTRIBUTING.md) と [`docs/product/DesignDoc.md`](docs/product/DesignDoc.md) を参照してください。

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
