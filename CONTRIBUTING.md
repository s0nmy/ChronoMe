# Contributing

## Code of Conduct

このリポジトリは学習プロジェクトです。相互に敬意を払い、建設的なコミュニケーションを心がけてください。

## Contribute 手順・ブランチ命名規則

1. Issue またはタスクを確認し、作業内容を明確にします。
2. `main` からブランチを作成します。
   - `feature/short-description`
   - `fix/short-description`
   - `chore/short-description`
   - `docs/short-description`
3. 変更をコミットし、PR を作成します。
4. 可能な限りテストと lint を実行します。

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

```bash
cd node-backend
npm install
npm run build
```

## テスト手順

```bash
cd backend
go test ./...
```

フロントエンドと Node API には現時点で自動テストがないため、変更時は lint の実行を推奨します。

## コード設計

- Go バックエンドは Handler / Usecase / Repository の層構造を維持します。
- React フロントエンドは `features/` 単位で機能を分割し、UI は `components/` に集約します。
- 仕様や設計の詳細は `docs/DesignDoc.md` を参照してください。
