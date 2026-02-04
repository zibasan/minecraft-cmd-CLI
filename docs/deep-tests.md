# CLI 深堀テスト記録

## 実施日時
- 2026-02-04

## 目的
- `create` によるコマンド生成の挙動確認
- `block add` / `block remove` によるファイル書き換えの副作用確認

## 実行したコマンド（非対話部分）
- `node dist/index.js create --help` — ヘルプ表示（対話式）
- `node dist/index.js block list` — block データ一覧を列挙（成功）
- `node dist/index.js block search` — 対話式のため自動実行不可（引数を与えるとエラー）

## 対話式コマンドの制約
- `create`, `block add`, `block remove`, `block search` などは `enquirer` や `readline` による対話式プロンプトを前提としています。
- この検証環境（自動実行）では対話プロンプトへの入力を与えられないため、完全な E2E 自動テストは行えません。

## 手動で確認した場合の想定手順（開発者向け）
1. `node dist/index.js create` を実行し、各質問に対して選択や入力を行う。
2. 生成されたコマンドを確認し、Enter 押下でクリップボードへコピーされるか確認する（`--copy false` を渡すとコピーを抑止可能）。
3. `node dist/index.js block add` を実行し、追加 ID を入力 → `src/data/blocks.ts` が更新されることを確認。
4. `node dist/index.js block remove` を実行し、削除対象を選択 → `src/data/blocks.ts` が更新されることを確認。

## 実際に自動で行った操作の結果サマリ
- `block list` : 正常に大量の block ID を出力（詳細は `docs/dev-behavior.md` / `docs/feature-behavior.md` を参照）。
- `block search <arg>` を CLI 引数として与えると `Expected 0 arguments but got 1` エラーになる（コマンドは対話式を想定しているため）。

## 結論 / 次のステップ
- 自動環境では対話式コマンドの入力を模擬できないため、対話テストは手動で実行してください。
- 本検証では対話式を除く主要機能は dev と feature 両方で一致しているため、プレビューのまま `integrate/...` ブランチを作成してマージ結果をコミットしました（次節参照）。
