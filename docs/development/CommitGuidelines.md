# Commit Guidelines

このドキュメントではコミットメッセージの書き方を整理します。

## プレフィックス

| Type | 意味 | 例 |
| --- | --- | --- |
| `feat` | 新機能追加 | `feat: add login button to header` |
| `fix` | バグ修正 | `fix: resolve crash on rotation` |
| `refactor` | 振る舞い変更なしの整理 | `refactor: simplify view logic` |
| `style` | フォーマット/見た目のみ | `style: format code with gofmt` |
| `docs` | ドキュメント修正 | `docs: update README usage` |
| `test` | テスト追加・変更 | `test: add unit test for report` |
| `chore` | 雑務 (ビルド/依存) | `chore: update deps` |
| `perf` | パフォーマンス改善 | `perf: optimize report aggregation` |
| `build` | ビルド/依存構成 | `build: upgrade tooling` |
| `ci` | CI/CD 設定 | `ci: add GitHub Actions workflow` |
| `revert` | 変更取り消し | `revert: revert "feat: add settings"` |

## スコープ付きコミット

`type(scope): message` の形式で対象領域を明示できます。

| Type | Scope | コミット例 | 意味 |
| --- | --- | --- | --- |
| `feat` | `ui` | `feat(ui): add dark mode toggle` | UI に新機能を追加 |
| `fix` | `api` | `fix(api): handle null response in login endpoint` | API 部分のバグ修正 |
| `refactor` | `auth` | `refactor(auth): simplify token validation` | 認証ロジックのリファクタ |
| `chore` | `deps` | `chore(deps): bump kotlin from 1.9.20 to 1.9.22` | 依存関係の更新 |
| `chore` | `release` | `chore(release): prepare v1.2.0` | リリース準備 |
| `ci` | `github` | `ci(github): fix build matrix` | CI 設定修正 |
| `test` | `api` | `test(api): add missing tests for auth` | API 部分のテスト追加 |
