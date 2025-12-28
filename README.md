# Redmine Exporter

Redmineのチケットを様々な形式（テキスト、Markdown、Excel）でエクスポートするツールです。VBA版とGo版の2つの実装があり、同じ設定ファイルを共有できます。

## 📁 リポジトリ構成

```
redmine-exporter/
├── vb/                     # VBA実装（Excel内で動作）
│   ├── main.vb            # VBAソースコード
│   └── README.md          # VBA版の詳細ドキュメント
├── go/                     # Go実装（CLIツール）
│   ├── cmd/               # エントリーポイント
│   ├── internal/          # 内部パッケージ
│   ├── Makefile           # ビルド設定
│   └── README.md          # Go版の詳細ドキュメント
├── redmine.config.sample   # 設定ファイルのサンプル（両方で共通）
├── .env.sample            # Docker環境変数のサンプル
├── docker-compose-redmine.yml  # ローカルRedmine環境
└── CLAUDE.md              # 開発者向けガイド
```

## 🚀 クイックスタート

### 1. 設定ファイルの準備

```bash
# サンプルをコピー
cp redmine.config.sample redmine.config

# エディタで編集
# - BaseUrl: RedmineのURL
# - ApiKey: あなたのAPIキー
# - FilterUrl: エクスポートするチケットのフィルタ条件
```

**設定例:**

```ini
[Redmine]
BaseUrl=https://your-redmine.example.com
ApiKey=your_api_key_here
FilterUrl=/issues.json?project_id=1&status_id=*&sort=parent:asc,id:asc

[TitleCleaning]
Pattern1=^\[.*?\]\s*      # [WIP]などのプレフィックスを削除
Pattern2=\s*\(.*?\)$      # (完了予定)などのサフィックスを削除
```

### 2. 実装を選択

#### 🔷 VBA版（Excel内で動作）

**特徴:**
- Excel内で直接実行
- Windowsのみ対応
- テーブル形式でExcelシートに出力

**使い方:**
```
詳細は vb/README.md を参照してください
```

#### 🔶 Go版（CLIツール）

**特徴:**
- スタンドアロンCLI
- クロスプラットフォーム（Linux/macOS/Windows）
- 複数の出力形式（Markdown/Text/Excel）

**インストール:**

```bash
cd go
make build
```

**実行:**

```bash
# Excel形式で出力
./bin/redmine-exporter -o output.xlsx

# Markdown形式で出力
./bin/redmine-exporter -o output.md

# テキスト形式で出力
./bin/redmine-exporter -o output.txt
```

詳細は `go/README.md` を参照してください。

## 📊 出力形式の比較

### テキスト形式（VBA版互換）

```
■親タスクA
・子タスクB 【進行中】 2026/01/02-2025/12/31 担当: 佐藤
⇒要約テキスト

・スタンドアロンタスク 【新規】 2026/01/15-2026/01/20 担当: 田中
⇒別の要約
```

### Markdown形式（Go版のみ）

```markdown
# 親タスクA

- **子タスクB** [進行中] 2026/01/02-2025/12/31 担当: 佐藤
  > 要約テキスト

- **スタンドアロンタスク** [新規] 2026/01/15-2026/01/20 担当: 田中
  > 別の要約
```

### Excel形式（両方対応）

| 親タスク | タスク名 | ステータス | 開始日 | 終了日 | 担当者 | 要約 |
|---------|---------|----------|--------|--------|--------|------|
| 親タスクA | 子タスクB | 進行中 | 2026/01/02 | 2025/12/31 | 佐藤 | 要約テキスト |
| - | スタンドアロンタスク | 新規 | 2026/01/15 | 2026/01/20 | 田中 | 別の要約 |

## 🐳 ローカルテスト環境（Docker）

Redmineをローカルで起動してテストできます。

### 環境変数の設定

```bash
cp .env.sample .env
# .envファイルを編集してパスワードなどを設定
```

### Redmineの起動

```bash
docker-compose -f docker-compose-redmine.yml up -d
```

### アクセス

- **URL**: http://localhost:3000 （ポート6000は使用しないでください - ブラウザでブロックされます）
- **初期ユーザー**: `admin`
- **初期パスワード**: `admin`

### 停止

```bash
docker-compose -f docker-compose-redmine.yml down
```

詳細は `vb/README.md` の「ローカルテスト環境」セクションを参照してください。

## 🔧 主要機能

### 共通機能（VBA版・Go版共通）

- **ページネーション対応**: 100件以上のチケットを自動で全件取得
- **タイトルクリーニング**: 正規表現で不要な文字列を削除
  - 例: `[WIP] タスク名` → `タスク名`
- **要約抽出**: チケット本文の `[要約]...[/要約]` タグを優先表示
- **親子関係**: 親チケットでグループ化して表示
- **スタンドアロンチケット**: 親を持たないチケットも出力可能

### タイトルクリーニング例

**設定:**
```ini
[TitleCleaning]
Pattern1=^\[.*?\]\s*      # プレフィックス削除
Pattern2=\s*\(.*?\)$      # サフィックス削除
Pattern3=【重要】          # 特定文字列削除
```

**変換例:**

| 元のタイトル | クリーニング後 |
|------------|-------------|
| `[WIP] ログイン機能の実装` | `ログイン機能の実装` |
| `バグ修正 (完了予定)` | `バグ修正` |
| `【重要】セキュリティ対応` | `セキュリティ対応` |

### 要約機能

チケット本文に `[要約]...[/要約]` タグを記述すると優先的に表示されます。

**例:**

```
[要約]ログイン画面のデザインを更新しました[/要約]

詳細:
- ロゴを新しいものに変更
- ボタンのスタイルを統一
- レスポンシブ対応を追加
```

出力される要約: `ログイン画面のデザインを更新しました`

タグがない場合は、本文の最初の非空行が表示されます。

## 📋 VBA版 vs Go版

| 項目 | VBA版 | Go版 |
|------|-------|------|
| **実行環境** | Excel内（Windows） | スタンドアロンCLI |
| **対応OS** | Windows | Linux, macOS, Windows |
| **出力形式** | テキスト、Excel | Markdown, テキスト, Excel |
| **出力先** | Excelセル・シート | ファイル |
| **セットアップ** | VBA-JSONライブラリが必要 | 単一バイナリ |
| **実行方法** | マクロ実行 | コマンドライン |
| **設定ファイル** | `redmine.config` (INI) | `redmine.config` (INI) ✅ 共通 |

### どちらを使うべきか？

- **VBA版を選ぶ場合:**
  - Excelで直接編集・加工したい
  - Windowsのみで使用
  - Excel環境が既にある

- **Go版を選ぶ場合:**
  - コマンドラインで自動化したい
  - macOSやLinuxで使用したい
  - Markdown形式で出力したい
  - Excel不要の軽量実行が必要

## 🔐 セキュリティ注意事項

- **APIキーの管理**: `redmine.config` にはAPIキーが含まれるため、Gitにコミットしないでください（`.gitignore`で除外済み）
- **環境変数ファイル**: `.env` ファイルもコミットしないでください
- **共有時の注意**: ファイルを共有する場合、`redmine.config` と `.env` は同梱しないでください
- **サンプルファイル**: `redmine.config.sample` と `.env.sample` をテンプレートとして使用してください

## 🛠️ 開発者向け

### テスト実行（Go版）

```bash
cd go
make test                 # 全テスト実行
make test-coverage       # カバレッジ付きテスト
```

### ビルド（Go版）

```bash
cd go
make build               # 単一プラットフォーム
make build-all           # クロスコンパイル（Linux/macOS/Windows）
```

### コードベース理解

プロジェクトのアーキテクチャや開発ガイドラインは `CLAUDE.md` を参照してください。

## 📖 ドキュメント

- **VBA版の詳細**: [vb/README.md](vb/README.md)
  - VBA-JSONライブラリのセットアップ
  - マクロの実行方法
  - トラブルシューティング

- **Go版の詳細**: [go/README.md](go/README.md)
  - ビルド方法
  - コマンドラインオプション
  - 開発・テスト方法

- **開発者ガイド**: [CLAUDE.md](CLAUDE.md)
  - アーキテクチャの詳細
  - テストパターン
  - 新機能の追加方法

## 🔗 参考リンク

- [Redmine REST API](https://www.redmine.org/projects/redmine/wiki/Rest_api)
- [VBA-JSON](https://github.com/VBA-tools/VBA-JSON)
- [Go言語公式サイト](https://golang.org/)

## 📄 ライセンス

MIT License

---

## トラブルシューティング

### エラー: "HTTP 401"

→ APIキーが間違っているか、有効期限が切れています。`redmine.config` の `ApiKey` を確認してください。

### エラー: "HTTP 406"

→ `FilterUrl` に `.json` が付いていません。URLの末尾に `.json` を追加してください。

**例:**
```ini
# ❌ 間違い
FilterUrl=/issues?project_id=1

# ✅ 正しい
FilterUrl=/issues.json?project_id=1
```

### エラー: "ERR_UNSAFE_PORT"（Docker使用時）

→ ポート6000はブラウザでブロックされています。`.env` ファイルで `REDMINE_PORT=3000` に変更するか、他の安全なポート（8080など）を使用してください。

### チケットが取得できない

→ `FilterUrl` の設定を確認してください。Redmineのチケット一覧画面で実際にフィルタリングして、そのURLをコピーすると確実です。

### Go版で "出力するチケットがありません"

→ フィルタ条件で取得されたチケットが1件もない可能性があります。`FilterUrl` の条件を確認してください。また、最新版では親を持たないチケット（スタンドアロンチケット）も出力されます。
