# Redmine Exporter

RedmineのチケットをExcelテーブル形式でエクスポートするVBAツールです。

---

## 機能

- **設定ファイル管理**: INI形式の設定ファイルで一括管理
- **フィルタURL対応**: Redmineのフィルタ結果URLを直接利用可能
- **ページネーション**: 100件以上のチケットも自動で全件取得
- **タイトルクリーニング**: 正規表現で不要な文字列を自動削除
- **要約抽出**: チケット本文の`[要約]...[/要約]`タグを優先表示
- **テーブル形式出力**: Excelの標準フィルタ機能で絞り込み可能

---

## 出力イメージ

Excelテーブル形式で出力されます。

| 親タスク | タスク名 | ステータス | 開始日 | 終了日 | 担当者 | 要約 |
|---------|---------|----------|--------|--------|--------|------|
| ユーザー認証機能 | ログイン画面作成 | 進行中 | 2025/01/02 | 2025/01/10 | 佐藤 | ログイン画面のUI実装 |
| ユーザー認証機能 | パスワード暗号化 | 完了 | 2025/01/05 | 2025/01/08 | 田中 | bcryptで暗号化 |

テーブル形式のため、Excelの標準フィルタ機能で自由に絞り込みができます。

## ローカルテスト環境（Docker）

テスト用にローカルでRedmineを立ち上げることができます。

### Redmineの起動

```bash
docker-compose -f docker-compose-redmine.yml up -d
```

### アクセス

- URL: http://localhost:3000
- 初期ユーザー: `admin`
- 初期パスワード: `admin`

### 初期設定

1. ブラウザで http://localhost:3000 にアクセス
2. `admin` / `admin` でログイン
3. パスワード変更を求められるので変更
4. 「管理」→「設定」→「API」で「RESTによるWebサービスを有効にする」をチェック
5. 右上のアカウント名→「個人設定」→「APIアクセスキー」で「表示」をクリック
6. 表示されたキーを`redmine.config`の`ApiKey`に設定

### プロジェクトとチケットの作成

テスト用にプロジェクトとチケットを作成します：

1. 「プロジェクト」→「新しいプロジェクト」
2. プロジェクト名を入力（例：「テストプロジェクト」）
3. プロジェクト作成後、URLから`project_id`を確認（例：`/projects/1`なら`id=1`）
4. 親チケットを作成（例：「ユーザー認証機能」）
5. 子チケットを作成し、親チケットを指定

### チケット本文に要約タグを追加

```
[要約]ログイン機能を実装しました[/要約]

詳細な説明...
```

### Redmineの停止

```bash
docker-compose -f docker-compose-redmine.yml down
```

データを完全に削除する場合：

```bash
docker-compose -f docker-compose-redmine.yml down -v
```

---

## セットアップ

### 1. 依存ライブラリのインストール

VBA-JSONライブラリが必要です。

1. [VBA-JSON](https://github.com/VBA-tools/VBA-JSON)から`JsonConverter.bas`をダウンロード
2. VBAエディタ（`Alt + F11`）で「ファイル」→「ファイルのインポート」から`JsonConverter.bas`をインポート

### 2. 設定ファイルの作成

`redmine.config.sample`を`redmine.config`にコピーして編集します。

**redmine.config の設定例:**

```ini
[Redmine]
BaseUrl=https://your-redmine.example.com
ApiKey=abc123def456...
FilterUrl=/issues.json?project_id=2&status_id=*&sort=parent:asc,id:asc

[TitleCleaning]
Pattern1=^\[.*?\]\s*
Pattern2=\s*\(.*?\)$
```

### 3. Redmine APIキーの取得

1. Redmineにログイン
2. 右上のアカウント名をクリック→「個人設定」
3. 「APIアクセスキー」セクションで「表示」をクリック
4. 表示されたキーを`redmine.config`の`ApiKey`に設定

### 4. フィルタURLの取得

1. Redmineのチケット一覧画面で任意の条件でフィルタリング
2. ブラウザのアドレスバーからURLをコピー
3. `/issues`以降の部分を`redmine.config`の`FilterUrl`に設定

**例:**
```
ブラウザのURL: https://redmine.example.com/issues?project_id=2&status_id=1&assigned_to_id=me
設定値: /issues.json?project_id=2&status_id=1&assigned_to_id=me
```

**注意:** URLの末尾に`.json`を付けることを忘れずに！

---

## 使い方

### 基本的な使い方

1. Excelファイルを開く
2. `Alt + F11`でVBAエディタを開く
3. `ExportRedmineSummary`マクロを実行

または、Excelシートにボタンを配置してマクロを割り当てることもできます。

### タイトルクリーニング

チケットタイトルから不要な文字列を自動削除できます。

**設定例:**

```ini
[TitleCleaning]
; [WIP]、[完了]などのプレフィックスを削除
Pattern1=^\[.*?\]\s*

; (完了予定)などのサフィックスを削除
Pattern2=\s*\(.*?\)$

; 特定の文字列を削除（例: "【重要】"）
Pattern3=【重要】
```

**適用例:**

| 元のタイトル | クリーニング後 |
|------------|-------------|
| `[WIP] ログイン機能の実装` | `ログイン機能の実装` |
| `バグ修正 (完了予定)` | `バグ修正` |
| `【重要】セキュリティ対応` | `セキュリティ対応` |

### 要約機能

チケット本文に`[要約]...[/要約]`タグを記述すると、その部分が優先的に表示されます。

**チケット本文の例:**

```
[要約]ログイン画面のデザインを更新しました[/要約]

詳細:
- ロゴを新しいものに変更
- ボタンのスタイルを統一
- レスポンシブ対応を追加

実装内容:
...（長文）...
```

Excelの「要約」列には「ログイン画面のデザインを更新しました」と表示されます。

タグがない場合は、本文の最初の非空行が表示されます。

---

## トラブルシューティング

### エラー: "HTTP 401"

→ APIキーが間違っているか、有効期限が切れています。`redmine.config`の`ApiKey`を確認してください。

### エラー: "設定ファイルが見つかりません"

→ `redmine.config`ファイルがExcelファイルと同じディレクトリにあるか確認してください。

### チケットが取得できない

→ `FilterUrl`の設定を確認してください。URLに`.json`が付いているか、クエリパラメータが正しいかチェックしてください。

### 正規表現が動作しない

→ VBScript.RegExpの構文に従っているか確認してください。パターンが不正な場合、そのパターンはスキップされます。

### JSONパースで落ちる

→ `JsonConverter.bas`が取り込まれていないか、VBA-JSONのバージョンが古い可能性があります。最新版をダウンロードしてください。

---

## 技術仕様

### ページネーション

100件以上のチケットがある場合、自動的にページネーションを実行して全件取得します。進捗は画面下部のステータスバーに表示されます。

### フィルタ条件

`redmine.config`の`FilterUrl`に設定したURLがそのまま使用されます。Redmineのチケット一覧で使用できるすべてのフィルタ条件が利用可能です。

**利用可能な主なパラメータ:**

- `project_id` - プロジェクトID
- `status_id` - ステータスID（`*`で全ステータス）
- `assigned_to_id` - 担当者ID（`me`で自分）
- `tracker_id` - トラッカーID
- `priority_id` - 優先度ID
- `sort` - ソート順（例: `parent:asc,id:asc`）

### データ構造

- 親チケット（`parent`がないチケット）ごとにグループ化
- 各親チケットの子チケットを一覧表示
- 親を持たないチケットは出力されません

---

## セキュリティ注意

- **APIキーの管理**: `redmine.config`ファイルにはAPIキーが含まれるため、Gitにコミットしないでください（`.gitignore`で除外済み）
- **共有時の注意**: Excelファイルを共有する場合、`redmine.config`は同梱しないでください
- **サンプルファイル**: `redmine.config.sample`をテンプレートとして使用してください

---

## ライセンス

MIT License

---

## 参考リンク

- [Redmine REST API](https://www.redmine.org/projects/redmine/wiki/Rest_api)
- [VBA-JSON](https://github.com/VBA-tools/VBA-JSON)
- [VBScript正規表現リファレンス](https://docs.microsoft.com/en-us/previous-versions/windows/internet-explorer/ie-developer/scripting-articles/ms974570(v=msdn.10))
