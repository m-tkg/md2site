# アーキテクチャと技術スタック

## 設計方針

md2site は当初 Hugo のラッパーとして構想されましたが、本質的な要件が「構造化ナビゲーション・ページ間移動・検索」の 3 点だと整理した結果、汎用 SSG を使わず必要な機能だけを自作する構成にしました。理由は次の通りです。

- Hugo をライブラリとして埋め込む方法は公式にサポートされておらず、API が不安定でバイナリも 80MB 超になる
- 任意のリポジトリの markdown は Hugo の front matter 規約を満たさないため、どのみち前処理層の自作が必要になる
- 要件 3 点は goldmark + `html/template` + 少量のクライアント JS で十分実現できる

結果として依存は 2 つだけで、バイナリは数 MB に収まっています。

## 技術スタック

| 技術 | 役割 |
|---|---|
| [goldmark](https://github.com/yuin/goldmark) | markdown → HTML 変換。Hugo 本体も採用している CommonMark 準拠パーサ。GFM 拡張（テーブル・打ち消し線・自動リンク等）を有効化 |
| [chroma](https://github.com/alecthomas/chroma) | コードブロックのシンタックスハイライト。ビルド時に CSS を生成し、ライト/ダーク両テーマを `prefers-color-scheme` で切替 |
| `html/template`（標準） | ページレイアウトの適用 |
| `embed.FS`（標準） | テーマ（レイアウト・CSS・JS）をバイナリに同梱 |
| `net/http`（標準） | `serve` サブコマンドのプレビューサーバ |

CLI フレームワーク（cobra 等）は使わず、標準の `flag` パッケージにサブコマンド分岐を組み合わせています。

## パッケージ構成

```
main.go
internal/
  cli/      フラグ解析、build / serve サブコマンド
  scan/     ディレクトリ走査と除外規則（node_modules 等の固定除外 + --exclude）
  site/     サイトモデル: Page、出力パス割当、ナビツリー構築
  render/   goldmark 変換、リンク書換、サイドバー HTML 生成
  search/   検索インデックス（search-index.js）の生成
  build/    パイプライン全体のオーケストレーションと出力ディレクトリ保護
  theme/    埋込テーマ（layout.html / style.css / app.js）と chroma CSS 生成
  server/   serve 用 HTTP サーバ
```

## ビルドパイプライン

1. **scan** — 入力ディレクトリを走査し、除外規則を適用して `.md` を収集
2. **site** — 各ファイルに出力パスを割当（`README.md` → `index.html`、`foo.md` → `foo/index.html`）。`foo.md` と `foo/README.md` の衝突は README 優先で解決
3. **parse** — 全ページを goldmark でパースし、最初の h1 をタイトルとして抽出
4. **nav** — ディレクトリ階層をミラーしたナビツリーを構築
5. **render** — ページごとに HTML を生成し、相対リンクを書換（後述）、レイアウトを適用
6. **search** — 各ページのプレーンテキストを抽出し検索インデックスを出力
7. **assets** — 参照されている画像だけを出力へコピー

## リンク書換の設計

リンク書換は goldmark の AST ではなく、**レンダリング後の HTML に対する src/href 属性の統一書換パス**として実装しています。README では `<p align="center"><img src="..."></p>` のような生 HTML が頻出し、これは AST 上では `Image` ノードにならず素通りしてしまうためです。HTML レベルで書き換えることで、markdown 記法のリンクと生 HTML を単一のコードパスで処理できます。

書換ルール:

- `foo.md` へのリンク → 変換後ページの URL（`#アンカー` は保持）
- ディレクトリへのリンク → そのディレクトリの README ページ
- 画像などの相対 src → 出力へコピーした上でパス書換
- 外部 URL・絶対パス・`mailto:` 等 → そのまま
- 存在しないファイルへのリンク → 警告を出してそのまま

生成される URL はすべて相対パス（`../../guide/index.html` 形式）です。このため base URL の設定が不要で、`file://` での直接閲覧と GitHub Pages のサブパス配信（`user.github.io/repo/`）の両方がそのまま動作します。
