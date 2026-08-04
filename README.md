# md2site

ディレクトリ（git リポジトリなど）内の markdown 群を、README を起点とした静的ドキュメントサイトに変換する CLI です。外部依存なしの単一バイナリで、構造化ナビゲーション・ページ間リンク・日本語対応の全文検索を備えた HTML を生成します。

## 特徴

- **単一バイナリ** — Hugo などの外部ツール不要。`go install` だけで使えます
- **README 起点** — 各ディレクトリの `README.md` がそのディレクトリの index ページになり、ルートの README がサイトトップになります
- **サイドバーナビ** — ディレクトリ階層をミラーした折りたたみ可能なツリー。現在ページをハイライト
- **日本語対応検索** — 正規化した部分一致検索（トークナイザ不要）。インデックスは `<script src>` で読み込むため `file://` で直接開いても動作します
- **リンク書換** — markdown 内の相対リンク（`foo.md`、`docs/`、生 HTML の `<img src>` 含む）を変換後の URL に書換。参照されている画像だけを出力へコピー
- **どこでも配信可能** — すべて相対リンクで生成するため、GitHub Pages のサブパス配下でも base URL 設定なしで動きます
- **コードハイライト** — chroma によるシンタックスハイライト（ライト/ダーク自動切替）

## ドキュメント

詳細は [docs/](docs/README.md) を参照してください（[公開版はこちら](https://m-tkg.github.io/md2site/)。md2site 自身で生成しています）。

- [CLI リファレンス](docs/cli.md)
- [アーキテクチャと技術スタック](docs/architecture.md)
- [検索の仕組み](docs/search.md)
- [GitHub Pages への公開](docs/github-pages.md)

## インストール

```sh
go install github.com/m-tkg/md2site@latest
```

アップデートは同じコマンドの再実行、または:

```sh
md2site upgrade    # 内部で go install ...@latest を実行
md2site version    # インストール中のバージョン確認
```

## 使い方

```sh
# ./public に生成
md2site build path/to/repo

# 出力先を指定
md2site build path/to/repo -o ./site

# ビルドしてローカルプレビュー
md2site serve path/to/repo --port 8080
```

### フラグ

| フラグ | 説明 |
|---|---|
| `-o <dir>` | 出力ディレクトリ（デフォルト `./public`） |
| `--exclude <glob>` | 追加の除外パターン。複数指定可（例: `--exclude 'CHANGELOG.md' --exclude 'drafts'`） |
| `--title <name>` | サイト名。省略時はルート README の h1 → 入力ディレクトリ名 |
| `--force` | マーカーのない空でない出力ディレクトリを強制上書き |
| `--port <n>` | serve のポート（デフォルト 8080） |

## 変換ルール

- 入力ディレクトリ以下の全 `.md` を再帰的に収集します。`node_modules/`・`vendor/`・`.git/`・隠しディレクトリは常に除外
- `README.md` → そのディレクトリの `index.html`、その他の `foo.md` → `foo/index.html`（pretty URLs）
- `foo.md` と `foo/README.md` が衝突する場合は README を優先し、`foo.md` は `foo.html` に出力して警告します
- ページタイトルは最初の `# 見出し`、なければファイル名
- 存在しないファイルへのリンクは警告を出してそのまま残します

## 出力ディレクトリの安全策

生成時に出力ディレクトリへ `.md2site` マーカーファイルを置きます。再ビルド時はマーカーがある場合のみ中身を削除して再生成します。マーカーのない空でないディレクトリを指定した場合はエラーで停止するため、誤って無関係なディレクトリを消すことはありません（意図的に上書きしたい場合は `--force`）。

## 開発

```sh
go test ./...
go build .
```
