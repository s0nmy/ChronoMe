
  # タイムカードWebアプリ

  This is a code bundle for タイムカードWebアプリ. The original project is available at https://www.figma.com/design/AbaLBz52XmmjUR7nq1bTXV/%E3%82%BF%E3%82%A4%E3%83%A0%E3%82%AB%E3%83%BC%E3%83%89Web%E3%82%A2%E3%83%97%E3%83%AA.

  ## Running the code

  Run `npm i` to install the dependencies.

  Run `npm run dev` to start the development server.


## フロントエンドの構成

- `src/App.tsx`: ログイン状態に応じた表示と画面の切り替え。
- `src/features/app/useChronome.ts`: 初期読み込み、認証、API を使った保存・削除と各機能の連携。
- `src/features/timer/`: 実行中タイマーの状態・操作と経過時間計算。
- `src/features/tags/useTags.ts`: 入力されたタグ名の解決と新規タグの作成。
- `src/features/entries/entryHelpers.ts`: エントリとプロジェクトの関連付け、タイトル生成。
- `src/components/`: 画面と UI コンポーネント。

保存済みエントリに表示するプロジェクト情報は、エントリ一覧と最新のプロジェクト一覧から導出します。プロジェクトの更新時に、エントリ側へ同じ情報を重ねて保存する必要はありません。

## 検証

```sh
npm run lint
npm test
npm run build
```

`npm test` は Node.js 22.18 以降の TypeScript 実行機能と組み込みテストランナーを使用します。経過時間計算、プロジェクト変更の反映、タイトル生成を検証します。API 通信やブラウザー操作のテストは含みません。
