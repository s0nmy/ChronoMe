# Time Allocation API

このドキュメントでは Go バックエンドに統合された Time Allocation API の仕様とアルゴリズムを説明します。

## 概要

- エンドポイント: `POST /api/allocations`
- 入力: 合計作業時間 `total_minutes` と、各タスクの `ratio` / 任意の `min_minutes` / `max_minutes`
- 出力: 各タスクの整数分配結果 (`allocated_minutes`)
- ストレージ: メイン DB に `allocation_requests` / `task_allocations` を永続化

## 入出力

### リクエスト

```jsonc
POST /api/allocations
{
  "total_minutes": 235,
  "tasks": [
    { "task_id": "a1", "ratio": 3, "min_minutes": 10, "max_minutes": 200 },
    { "task_id": "b2", "ratio": 2 },
    { "task_id": "c3", "ratio": 1 }
  ]
}
```

- `total_minutes`: 1 以上の整数
- `tasks`: 1 件以上。`task_id` はユニークで、`ratio` は正数。`min_minutes`/`max_minutes` は任意 (整数)。

### レスポンス (201)

```json
{
  "request_id": "3e0a5e1e-8045-493d-86f2-777f518146d3",
  "total_minutes": 235,
  "allocations": [
    { "task_id": "a1", "ratio": 3, "allocated_minutes": 118 },
    { "task_id": "b2", "ratio": 2, "allocated_minutes": 78 },
    { "task_id": "c3", "ratio": 1, "allocated_minutes": 39 }
  ]
}
```

エラー時は:

- 422: バリデーション / 制約違反 (`{ "errors": ... }` / `{ "error": "..." }`)
- 500: 想定外エラー

## バリデーション

1. `total_minutes > 0`
2. `tasks.length >= 1`
3. `ratio` は正数 (`> 0`)
4. `task_id` はユニーク
5. `min_minutes` / `max_minutes` は整数（`min <= max`）
6. `sum(min_minutes)` が `total_minutes` を超えない
7. すべてのタスクに `max` が定義されている場合のみ、`sum(max_minutes)` が `total_minutes` 以上であることを確認

## 分配アルゴリズム

1. **正規化**: `normalizedRatio = ratio / sum(ratios)`。
2. **最小値の確保**: 各タスクに `min_minutes` を事前配分。残り時間を `remaining` とする。
3. **基礎割当**: `remaining * normalizedRatio` を計算し `floor` で整数化。タスクに `max` があれば上限までに制限。余り (`remainder`) を保持。
4. **端数調整 (最大剰余法)**:
   - `remaining` が 0 になるまで、`remainder` の大きい順に 1 分ずつ配分。
   - `max` に達したタスクはスキップ。全タスクが `max` に到達して残りがある場合は 422 を返す。
   - 対象が 1 つだけの場合は残りを一括配分。
5. **結果整合性**: 常に `sum(allocated_minutes) === total_minutes`。

## ストレージ仕様

```
allocation_requests(id TEXT PK, total_minutes INTEGER, created_at TEXT)
task_allocations(
  id INTEGER PK AUTOINCREMENT,
  request_id TEXT FK,
  task_id TEXT,
  ratio REAL,
  allocated_minutes INTEGER,
  min_minutes INTEGER NULL,
  max_minutes INTEGER NULL,
  created_at TEXT,
  updated_at TEXT
)
```

1 回の API 呼び出しにつき 1 行の `allocation_requests` と複数行の `task_allocations` をトランザクションで登録します。

## 使い方

```
cd backend
go run ./cmd/server
```

テーブルは Go サーバー起動時の自動マイグレーションで作成されます。
