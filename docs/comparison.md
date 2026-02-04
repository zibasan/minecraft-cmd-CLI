# dev と tukuyomi/feature/changes-command-behavior の比較

以下は、`dev` と `tukuyomi/feature/changes-command-behavior` ブランチで実行・確認したコマンドと挙動の比較です。

| コマンド | dev に存在 | feature に存在 | 挙動（dev） | 挙動（feature） | 備考 |
|---|---:|---:|---|---|---|
| `mccmd help` | ✅ | ✅ | CLI ヘルプを表示 | 同上 | 同一出力 |
| `mccmd create` | ✅ | ✅ | `create` のヘルプのみ（サブコマンドなし） | 同上 | 同一 |
| `mccmd block` | ✅ | ✅ | `block` のサブコマンドあり: `add, list, remove, search` | 同上 | 同一 |
| `mccmd block list` | ✅ | ✅ | block ID を大量に列挙（正常） | 同上 | 同一 |
| `mccmd block search <arg>` | ✅ | ✅ | 引数を与えると `too many arguments` エラー（Expected 0 arguments but got 1） | 同上 | 両ブランチで同じ挙動。仕様変更がある場合はここを要確認 |

結論: テストした範囲では、両ブランチの CLI 表示・主要サブコマンドは同一でした。唯一目立った点は `block search` の引数仕様で、実際に引数を与えるとエラーになりますが、これは両ブランチとも同様の挙動でした。

次は、実際にマージした場合のプレビュー検証（`git merge --no-commit --no-ff`）を行い、コンフリクトや動作差分がないか確認します。
