# tukuyomi/feature/changes-command-behavior ブランチでの挙動記録

## ブランチ情報
- チェックアウト: `tukuyomi/feature/changes-command-behavior`

## 実行日時
- 2026-02-04

## 実行したコマンド
- `git checkout tukuyomi/feature/changes-command-behavior`
- `node dist/index.js help`
- `node dist/index.js create --help`
- `node dist/index.js block --help`
- `node dist/index.js block list`
- `node dist/index.js block search stone` (テスト — 引数を与えるとエラーになる)

## `mccmd` ヘルプ出力（抜粋）
Commands:
  create          Generate Minecraft commands
  block           Manage block ID definitions
  help [command]  display help for command

## `block` サブコマンド（抜粋）
- `add`    Add a custom block ID to src/data/blocks.ts/json
- `list`   List known block IDs
- `remove` Remove a block ID from src/data/blocks.ts/json
- `search` Search blocks by category and name

## 実行ログ抜粋 / 備考
- `node dist/index.js block list` は正常に block ID を列挙しました。
- `node dist/index.js block search stone` を実行したところ、`error: too many arguments for 'search'. Expected 0 arguments but got 1.` が出力され、引数付きの実行は失敗しました。

---
記録終了。
