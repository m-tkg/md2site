# GitHub Pages への公開

md2site で生成したサイトは GitHub Pages で無料公開できます。この手順書は、md 更新を push するたびにサイトが自動更新される GitHub Actions 方式を中心に説明します。

## GitHub Pages とは

- GitHub が提供する静的サイトホスティング。**public リポジトリなら無料**（private リポジトリでの利用は有料プラン）
- URL は `https://<ユーザー名>.github.io/<リポジトリ名>/`
- 制限: サイト容量 1GB、帯域 目安 100GB/月、ビルド 10 回/時 — 個人のドキュメントサイトでは実質問題になりません
- EC サイトなどの商用利用は規約上不可。ドキュメント公開は問題ありません

md2site はすべてのリンクを相対パスで生成するため、`/<リポジトリ名>/` のサブパス配下でも **base URL の設定なし**でそのまま動きます。

## 方法 1: GitHub Actions で自動デプロイ（推奨）

push のたびに Actions がサイトを再生成して公開します。このリポジトリ自身が使っている workflow は次の通りです（`.github/workflows/pages.yml`）。

```yaml
name: Deploy docs to GitHub Pages

on:
  push:
    branches: [main]
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: true

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version: stable
      - run: go run . build . -o _site --exclude internal
      - uses: actions/upload-pages-artifact@v5
        with:
          path: _site

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - id: deployment
        uses: actions/deploy-pages@v5
```

他のリポジトリで使う場合は build ステップを次のように変えます。

```yaml
      - run: go run github.com/m-tkg/md2site@latest build . -o _site
```

### 有効化手順

1. workflow ファイルを配置して push
2. リポジトリの **Settings → Pages → Source** を **GitHub Actions** に設定（CLI なら `gh api -X POST repos/<owner>/<repo>/pages -f build_type=workflow`）
3. Actions の実行完了後、`https://<ユーザー名>.github.io/<リポジトリ名>/` で公開

## 方法 2: ブランチ配信（手動）

Actions を使わず、生成した HTML をブランチに push する方法です。

```sh
md2site build . -o /tmp/site
git checkout --orphan gh-pages
git rm -rf .
cp -r /tmp/site/* .
touch .nojekyll        # Jekyll 処理を無効化（アンダースコア始まりのパス対策）
git add -A && git commit -m "deploy" && git push origin gh-pages
```

Settings → Pages → Source を「Deploy from a branch」にし、`gh-pages` / `/ (root)` を指定します。手軽ですが更新のたびに手作業になるため、継続運用には方法 1 を推奨します。

## トラブルシューティング

- **404 になる** — Actions の deploy ジョブが成功しているか、Settings → Pages の Source が GitHub Actions になっているかを確認
- **CSS が当たらない・リンク切れ** — md2site 生成物なら起きない構成ですが、他ツール併用時はサブパス（`/<リポジトリ名>/`）を絶対パス `/...` で参照していないか確認
- **反映が遅い** — CDN キャッシュで数分かかることがあります
