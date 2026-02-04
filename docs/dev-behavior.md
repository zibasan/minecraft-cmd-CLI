# dev ブランチでの挙動記録

## ブランチ情報
- チェックアウト: `dev` (origin/dev と同期)

## 実行日時
- 2026-02-04

## 実行したコマンド
- `git checkout dev`
- `node dist/index.js help`
- `node dist/index.js create --help`
- `node dist/index.js block --help`
- `node dist/index.js block list`
- `node dist/index.js block search stone` (テスト — 引数を与えるとエラーになる)

## `mccmd` ヘルプ出力
Usage: mccmd [options] [command]

Generate Minecraft Java Edition command on CLI.

Options:
  -V, --version   output the version number
  -h, --help      display help for command

Commands:
  create          Generate Minecraft commands
  block           Manage block ID definitions
  help [command]  display help for command

## `create` コマンド (ヘルプ)
Usage: mccmd create [options]

Generate Minecraft commands

Options:
  -h, --help  display help for command

## `block` コマンド (ヘルプ)
Usage: mccmd block [options] [command]

Manage block ID definitions

Subcommands:
- `add`    Add a custom block ID to src/data/blocks.ts/json
- `list`   List known block IDs
- `remove` Remove a block ID from src/data/blocks.ts/json
- `search` Search blocks by category and name
- `help`   display help for command

## 実行ログ抜粋 / 備考
- `node dist/index.js block list` は大量の block ID を標準出力に列挙しました（省略していますが正常に一覧が出ました）。
- `node dist/index.js block search stone` を実行したところ、`error: too many arguments for 'search'. Expected 0 arguments but got 1.` が出力されました。`search` の引数仕様が dev と feature ブランチで変わっている可能性があります。

---
記録終了。
