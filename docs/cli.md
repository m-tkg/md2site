# CLI リファレンス

## コマンド

### `md2site build <入力dir> [フラグ]`

入力ディレクトリの markdown から静的サイトを生成します。

```sh
md2site build path/to/repo              # ./public に生成
md2site build path/to/repo -o ./site    # 出力先を指定
```

### `md2site serve <入力dir> [フラグ]`

一時ディレクトリにビルドし、ローカル HTTP サーバでプレビューします。

```sh
md2site serve path/to/repo --port 8080
# → http://127.0.0.1:8080/
```

## フラグ

| フラグ | 対象 | 説明 |
|---|---|---|
| `-o <dir>` | build | 出力ディレクトリ。デフォルト `./public` |
| `--exclude <glob>` | 共通 | 追加の除外パターン。複数指定可 |
| `--title <name>` | 共通 | サイト名。省略時はルート README の h1 → 入力ディレクトリ名 |
| `--force` | build | マーカーのない空でない出力ディレクトリを強制上書き |
| `--port <n>` | serve | ポート番号。デフォルト 8080 |

フラグは位置引数の前後どちらに書いても解釈されます。

## 対象ファイルと除外規則

入力ディレクトリ以下の全 `.md` を再帰的に収集します。以下は常に除外されます。

- `node_modules/`、`vendor/`、`.git/`
- 隠しディレクトリ・隠しファイル（`.` 始まり）

`--exclude` のパターンは次のいずれかにマッチすると除外されます。

- 相対パス全体への glob マッチ（例: `--exclude 'docs/drafts/*.md'`）
- ファイル名・ディレクトリ名への glob マッチ（例: `--exclude '*_test.md'`）
- サブツリー指定（例: `--exclude drafts` で `drafts/` 以下すべて）

```sh
md2site build . --exclude CHANGELOG.md --exclude internal
```

## 変換ルール

| 入力 | 出力 |
|---|---|
| `README.md` | `index.html`（サイトトップ） |
| `docs/README.md` | `docs/index.html` |
| `docs/setup.md` | `docs/setup/index.html` |

- ページタイトルは最初の `# 見出し`。なければファイル名
- サイドバーはディレクトリ階層をミラーし、README を持つディレクトリはラベル自体がリンクになります
- `foo.md` と `foo/README.md` が両方ある場合は README が `foo/index.html` を取り、`foo.md` は `foo.html` に出力されます（警告表示）

## 出力ディレクトリの保護（マーカー方式）

build は出力ディレクトリ直下に `.md2site` というマーカーファイルを書き込みます。再ビルド時の挙動:

- マーカーあり → 中身を全削除して再生成（安全: 自分が生成したものだけを消す）
- 空ディレクトリ / 存在しない → そのまま生成
- **マーカーなしで中身あり → エラー終了**。誤って無関係なディレクトリを指定してもファイルは消えません。意図的に上書きする場合のみ `--force`
