# Commit Message Style Guide

このプロジェクトでは、コミットメッセージの統一と履歴の可読性向上のため、**Conventional Commits 1.0.0** に準拠します。

## 概要

Conventional Commits は、コミットメッセージに一定の構造を持たせるための軽量な規約です。
変更内容を `type` で明示することで、履歴の把握、CHANGELOG 生成、リリース自動化、Semantic Versioning との連携がしやすくなります。

## 基本フォーマット

```text
<type>[optional scope][!]: <title>

[optional body]

[optional footer(s)]
```

例:

```text
feat(api): 商品出荷時にメール送信を追加
```

## 必須ルール

- コミットは `type` で始める
- `type` の後ろに任意で `scope` を付けられる
- 破壊的変更がある場合は `!` を付けられる
- `type` / `scope` の後ろには必ず `: ` を入れる
- `title` は変更内容の短い要約にする
- `body` を書く場合は、タイトルの後に 1 行空ける
- `footer` を書く場合は、本文の後に 1 行空ける

## 主な type

| type | 意味 | SemVer への影響 |
|------|------|----------------|
| feat | 新機能の追加 | MINOR |
| fix | バグ修正 | PATCH |
| docs | ドキュメント変更 | なし |
| refactor | 挙動を変えない構造改善 | なし |
| perf | パフォーマンス改善 | なし |
| test | テスト追加・修正 | なし |
| build | ビルド関連の変更 | なし |
| ci | CI/CD 設定変更 | なし |
| chore | 雑多な保守作業 | なし |
| style | フォーマットなどの非機能変更 | なし |
| revert | 変更の取り消し | 状況による |

`feat` と `fix` 以外の type も利用できますが、**BREAKING CHANGE を含まない限り SemVer 上の明確な意味は持ちません**。

## scope

`scope` は変更対象の文脈を補足する任意情報です。括弧で囲みます。

例:

```text
feat(ios): アニメーションを追加
fix(parser): 配列解析時の空白処理を修正
docs(readme): セットアップ手順を追記
```

想定スコープ例:

- `ios`
- `android`
- `backend`
- `db`
- `infra`
- `readme`

## 破壊的変更

API や互換性を壊す変更は、次のいずれかで明示します。

### 1. `!` を付ける

```text
feat!: Node 6 のサポートを終了
feat(api)!: 商品出荷時にメール送信を追加
```

### 2. `BREAKING CHANGE:` フッターを書く

```text
feat: 設定オブジェクトの拡張を許可

BREAKING CHANGE: config の extends キーは他の設定ファイルの継承に使用される
```

`!` と `BREAKING CHANGE:` は併用しても構いません。  
破壊的変更を含むコミットは **MAJOR** リリース相当です。

## body / footer

詳細説明が必要な場合のみ `body` を使います。

差分が大きいコミットでは、`title` だけで終わらせず、`body` に変更の**概要と詳細**を書いてください。  
その際、詳細はファイル単位の列挙ではなく、**機能単位のまとまり**で箇条書きにしてください。  
実装ファイル名や作業ログではなく、レビュアーが変更内容を把握しやすい粒度で記述します。  
内部実装の部品ごとに分解しすぎず、**ユーザーに見える振る舞い**や**変更の目的**が伝わる単位でまとめてください。
変更区分が少なく、`title` だけで十分に意図が伝わる場合は、`body` を省略して構いません。  
`body` が必要な場合でも、まずは 1 から 2 点で足りるかを優先し、細かな実装差分を並べすぎないようにしてください。

細かすぎる例:

```text
feat(encounter): モーフィングトランジションの実装と環境キーの追加

* EncounterNamespaceKey を追加し、EnvironmentValues に統合
* EncounterDetailView にモーフィングトランジション機能を追加
* EncounterListView に選択されたエンカウンターの詳細表示機能を実装
* EncounterRow でのマッチした要素の非表示機能を実装
```

望ましい例:

```text
refactor(ios): 詳細画面への遷移体験を改善
```

必要な場合でも、この程度に留めます:

```text
refactor(ios): 詳細画面への遷移体験を改善

* 一覧から詳細への遷移アニメーションを見直し、画面遷移を自然にした
* 遷移中の表示崩れを抑え、見た目の一貫性を改善した
```

例:

```text
feat(infra): Terraform による GCP 基盤を構築

* Terraform state を GCS backend で管理する構成を追加
* GCP 利用に必要な API をまとめて有効化
* Cloud Run / Scheduler / CI 向けのサービスアカウントと WIF を整備
* アプリ配布用の Artifact Registry を作成
* GitHub Actions から terraform plan / apply を実行する CI/CD を追加
* Terraform 運用に必要な ignore 設定を追加
```

フッターは `token: value` 形式を基本とし、`Acked-by` のように単語区切りは `-` を使います。  
例外として `BREAKING CHANGE:` はそのまま使用できます。

## この規約を使う理由

- CHANGELOG を自動生成しやすい
- リリース種別を自動判定しやすい
- 変更の意図をチームや利用者に伝えやすい
- 履歴の検索性と保守性が上がる
- ビルドや公開フローの自動化に繋げやすい

## このプロジェクトでの運用ルール

- コミットメッセージは **日本語** で書く
  - ただし英語の単語は英語を保持する
- タイトルは **簡潔に 50 文字以内を目安** とする
- タイトル末尾に句読点は付けない
- 1 コミット 1 意図を基本にする
- 複数の性質が混ざる場合は、できるだけコミットを分割する
