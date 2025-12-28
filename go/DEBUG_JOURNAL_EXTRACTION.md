# ジャーナルからのタグ抽出デバッグガイド

## 問題

`--include-comments` フラグを使用しても、コメント内のタグ（例: `[進捗]バグが発生しました[/進捗]`）が抽出されないという報告がありました。

## 追加したデバッグ出力

デバッグ用のログ出力を追加しました。以下の情報が表示されます：

1. **needsJournals フラグの状態** (main.go:272)
   - `needsJournals` が true になっているか
   - `cfg.Output.IncludeComments` の値
   - `commentsMode` の値

2. **取得したジャーナルの統計** (main.go:288-302)
   - 取得したジャーナルの合計件数
   - Notes（コメント本文）が空でないジャーナルの件数

## テスト方法

### 1. デバッグビルドの実行

```bash
# デバッグビルドを作成（既に作成済み）
go build -o ./bin/redmine-exporter-debug ./cmd/redmine-exporter

# デバッグ版で実行
./bin/redmine-exporter-debug \
  -o /tmp/test_output.md \
  --mode tags \
  --tags "進捗,課題,要約" \
  --include-comments \
  2>&1 | grep DEBUG
```

### 2. 期待される出力

正常に動作している場合、以下のようなデバッグメッセージが表示されます：

```
[DEBUG] needsJournals=true (IncludeComments=true, mode=)
[DEBUG] 取得したジャーナル: 合計50件 (Notes有り: 30件)
```

### 3. 確認ポイント

#### ケース1: needsJournals が false の場合
```
[DEBUG] needsJournals=false (IncludeComments=false, mode=)
```
→ `--include-comments` フラグが正しく解釈されていない
→ 設定ファイルの `IncludeComments` が false に設定されている可能性

#### ケース2: needsJournals は true だがジャーナルが 0 件の場合
```
[DEBUG] needsJournals=true (IncludeComments=true, mode=)
[DEBUG] 取得したジャーナル: 合計0件 (Notes有り: 0件)
```
→ Redmine API が journals を返していない
→ `include=journals` パラメータが正しく送信されていない可能性
→ Redmine 側の権限設定の問題の可能性

#### ケース3: ジャーナルはあるが Notes が空の場合
```
[DEBUG] needsJournals=true (IncludeComments=true, mode=)
[DEBUG] 取得したジャーナル: 合計50件 (Notes有り: 0件)
```
→ ジャーナルは取得されているが、コメント本文（Notes）が空
→ 変更履歴のみでコメントが付いていない可能性

#### ケース4: ジャーナルもNotes もあるのにタグが抽出されない場合
```
[DEBUG] needsJournals=true (IncludeComments=true, mode=)
[DEBUG] 取得したジャーナル: 合計50件 (Notes有り: 30件)
```
→ Processor の `includeComments` が false になっている可能性
→ タグの形式が正しくない（例: 全角の`[`でなく半角の`[`を使用）

## 詳細デバッグ（必要に応じて）

さらに詳細なデバッグが必要な場合、以下のコメントアウトされたデバッグ行を有効化してください：

### processor.go の詳細ログ

**ファイル**: `internal/processor/processor.go`

```go
// 行 75-77 をコメント解除
if len(issue.ExtractedTags) > 0 {
    fmt.Fprintf(os.Stderr, "[DEBUG] Issue #%d: ExtractedTags=%v (journals=%d)\n",
        issue.ID, issue.ExtractedTags, len(issue.Journals))
}

// 行 163 をコメント解除
fmt.Fprintf(os.Stderr, "[DEBUG] ExtractTags: includeComments=true, journals=%d\n", len(journals))

// 行 175 をコメント解除
fmt.Fprintf(os.Stderr, "[DEBUG] ExtractTags: 抽出成功 tagName=%s, content=%s\n", tagName, content)
```

これにより、以下の情報が表示されます：
- どのチケットからタグが抽出されたか
- 各チケットが持っているジャーナルの数
- どのタグが抽出されたか

## ユニットテストの確認

タグ抽出ロジック自体は正しく動作することを確認済み：

```bash
go test -v ./internal/processor -run "TestExtractTags_FromJournals|TestProcess_WithJournalTags"
```

出力:
```
=== RUN   TestExtractTags_FromJournals
=== RUN   TestExtractTags_FromJournals/ジャーナルからタグ抽出（includeComments=true）
=== RUN   TestExtractTags_FromJournals/ジャーナルからタグ抽出（includeComments=false）
=== RUN   TestExtractTags_FromJournals/説明文とジャーナル両方にタグ（説明文優先）
=== RUN   TestExtractTags_FromJournals/空のジャーナル
=== RUN   TestExtractTags_FromJournals/複数のジャーナル（最初のタグのみ抽出）
--- PASS: TestExtractTags_FromJournals (0.00s)
=== RUN   TestProcess_WithJournalTags
--- PASS: TestProcess_WithJournalTags (0.00s)
PASS
```

## 次のステップ

1. デバッグ版を実行してデバッグ出力を確認
2. 上記の「確認ポイント」を参照して問題を特定
3. 必要に応じて詳細デバッグログを有効化
4. 実際の Redmine チケットでタグが正しい形式になっているか確認

## よくある問題

### タグの形式が間違っている

コメント内のタグは以下の形式である必要があります：

**正しい**:
```
[進捗]バグが発生しました[/進捗]
```

**間違い**:
```
【進捗】バグが発生しました【/進捗】  # 全角カッコが違う
[進捗］バグが発生しました［/進捗]    # 半角・全角が混在
[進捗 ]バグが発生しました[ /進捗]    # スペースが入っている
```

### 設定ファイルで IncludeComments が false に設定されている

`redmine_exporter.ini` の内容を確認：

```ini
[Output]
IncludeComments=false  # これを true にするか、フラグ --include-comments を使用
```

### Redmine の権限設定

ユーザーにジャーナル（コメント）の閲覧権限がない場合、API がジャーナルを返しません。
Redmine の管理画面でユーザーのロールと権限を確認してください。
