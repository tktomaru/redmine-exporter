# Redmine Exporter (Go版)

RedmineのチケットをMarkdown、テキスト、Excel形式でエクスポートするCLIツールです。

## 特徴

- **VBA版との互換性**: 同じ設定ファイル（redmine.config）を使用可能
- **複数の出力形式**: Markdown (.md), テキスト (.txt), Excel (.xlsx)
- **クロスプラットフォーム**: Linux、macOS、Windows対応
- **高速**: Go言語による高速な処理
- **スタンドアロン**: 単一バイナリで動作

## インストール

### ビルドから

```bash
cd go
make build
```

バイナリは `bin/redmine-exporter` に生成されます。

### クロスコンパイル

```bash
make build-all
```

以下のバイナリが生成されます：
- `bin/redmine-exporter-linux-amd64`
- `bin/redmine-exporter-darwin-amd64`
- `bin/redmine-exporter-windows-amd64.exe`

## 使い方

### 基本的な使い方

```bash
# Excel形式で出力
./bin/redmine-exporter -o output.xlsx

# Markdown形式で出力
./bin/redmine-exporter -o output.md

# テキスト形式で出力
./bin/redmine-exporter -o output.txt
```

### 設定ファイルを指定

```bash
./bin/redmine-exporter -c custom.config -o output.xlsx
```

### ヘルプ表示

```bash
./bin/redmine-exporter -h
```

### バージョン表示

```bash
./bin/redmine-exporter -v
```

## 設定ファイル

VBA版と同じ `redmine.config` を使用します。プロジェクトルートの `redmine.config.sample` を参考にしてください。

```ini
[Redmine]
BaseUrl=https://redmine.example.com
ApiKey=YOUR_API_KEY
FilterUrl=/issues.json?project_id=1&status_id=*&sort=parent:asc,id:asc

[TitleCleaning]
Pattern1=^\[.*?\]\s*
Pattern2=\s*\(.*?\)$
```

## 出力形式

### Markdown形式

```markdown
# 親タスクA

- **タスクB** [進行中] 2026/01/02-2025/12/31 担当: 佐藤
  > ひとこと整形で変更しました
```

### テキスト形式（VBA版互換）

```
■親タスクA
・タスクB 【進行中】 2026/01/02-2025/12/31 担当: 佐藤
⇒ひとこと整形で変更しました
```

### Excel形式（VBA版互換）

テーブル形式で出力されます：

| 親タスク | タスク名 | ステータス | 開始日 | 終了日 | 担当者 | 要約 |
|---------|---------|----------|--------|--------|--------|------|

## 開発

### テスト実行

```bash
make test
```

### カバレッジ確認

```bash
make test-coverage
```

### クリーンアップ

```bash
make clean
```

## VBA版との違い

| 機能 | VBA版 | Go版 |
|------|-------|------|
| 実行環境 | Excel内 | スタンドアロンCLI |
| 出力先 | Excelセル | ファイル |
| 出力形式 | テキスト、Excel | Markdown、テキスト、Excel |
| プラットフォーム | Windows | Linux、macOS、Windows |
| 設定ファイル | redmine.config (INI) | 同じ |

## トラブルシューティング

### エラー: "設定ファイルの読み込みに失敗"

→ `redmine.config` がカレントディレクトリまたは `-c` で指定したパスに存在することを確認してください。

### エラー: "HTTP 401"

→ APIキーが正しいか確認してください。

### エラー: "未対応の拡張子"

→ 出力ファイルの拡張子は `.md`, `.txt`, `.xlsx` のいずれかを使用してください。

## ライセンス

MIT License
