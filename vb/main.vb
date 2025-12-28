Option Explicit

' ===== 設定 =====
Private Const REDMINE_BASE As String = "https://redmine.example.com"
Private Const API_KEY As String = "YOUR_API_KEY"
Private Const PROJECT_ID As Long = 2

Public Sub ExportRedmineSummary()
    ' 設定ファイルから値を読み込み
    Dim baseUrl As String
    Dim apiKey As String
    Dim filterUrl As String

    baseUrl = GetConfigValue("Redmine", "BaseUrl", REDMINE_BASE)
    apiKey = GetConfigValue("Redmine", "ApiKey", API_KEY)
    filterUrl = GetConfigValue("Redmine", "FilterUrl", "/issues.json?project_id=" & PROJECT_ID & "&status_id=*&sort=parent:asc,id:asc")

    ' 全チケットを取得（ページネーション対応）
    Dim issues As Collection
    Set issues = FetchAllIssues(baseUrl, filterUrl, apiKey)

    ' 親子関係を構築
    Dim byId As Object
    Dim childrenByParent As Object
    Dim roots As Collection

    Set byId = CreateObject("Scripting.Dictionary")
    Set childrenByParent = CreateObject("Scripting.Dictionary")
    Set roots = New Collection

    Dim it As Variant
    For Each it In issues
        byId(CLng(it("id"))) = it
    Next

    For Each it In issues
        If HasKey(it, "parent") Then
            Dim pid As Long
            pid = CLng(it("parent")("id"))
            If Not childrenByParent.Exists(pid) Then
                Dim col As New Collection
                childrenByParent(pid) = col
            End If
            childrenByParent(pid).Add it
        Else
            roots.Add it
        End If
    Next

    ' テーブル形式で出力
    OutputToTable byId, childrenByParent, roots
End Sub

Private Function HttpGetJson(ByVal url As String, ByVal apiKey As String) As String
    Dim http As Object
    Set http = CreateObject("MSXML2.ServerXMLHTTP.6.0")

    http.Open "GET", url, False
    http.setRequestHeader "X-Redmine-API-Key", apiKey
    http.setRequestHeader "Accept", "application/json"
    http.Send

    If http.Status < 200 Or http.Status >= 300 Then
        Err.Raise vbObjectError + 1, , "HTTP " & http.Status & ": " & http.responseText
    End If

    HttpGetJson = http.responseText
End Function

Private Function FetchAllIssues(ByVal baseUrl As String, ByVal filterUrl As String, ByVal apiKey As String) As Collection
    Dim allIssues As New Collection
    Dim offset As Long
    Dim limit As Long
    Dim totalCount As Long

    offset = 0
    limit = 100
    totalCount = -1

    Do
        ' URLにlimitとoffsetを追加
        Dim url As String
        url = baseUrl & filterUrl

        ' フィルタURLに既にクエリパラメータが含まれているか確認
        If InStr(filterUrl, "?") > 0 Then
            url = url & "&limit=" & limit & "&offset=" & offset
        Else
            url = url & "?limit=" & limit & "&offset=" & offset
        End If

        ' 進捗表示
        If totalCount > 0 Then
            Application.StatusBar = "Redmineからチケットを取得中... (" & allIssues.Count & " / " & totalCount & ")"
        Else
            Application.StatusBar = "Redmineからチケットを取得中..."
        End If

        Dim jsonText As String
        jsonText = HttpGetJson(url, apiKey)

        Dim root As Object
        Set root = JsonConverter.ParseJson(jsonText)

        ' 初回のみtotal_countを取得
        If totalCount = -1 Then
            totalCount = CLng(root("total_count"))
        End If

        Dim issues As Collection
        Set issues = root("issues")

        Dim it As Variant
        For Each it In issues
            allIssues.Add it
        Next

        offset = offset + limit

        ' 全件取得完了したらループを抜ける
        If allIssues.Count >= totalCount Then
            Exit Do
        End If
    Loop

    ' 進捗表示をクリア
    Application.StatusBar = False

    Set FetchAllIssues = allIssues
End Function

Private Function HasKey(ByVal dict As Object, ByVal key As String) As Boolean
    On Error GoTo EH
    Dim tmp As Variant
    tmp = dict(key)
    HasKey = True
    Exit Function
EH:
    HasKey = False
End Function

Private Function Nz(ByVal v As Variant, ByVal fallback As String) As String
    If IsObject(v) Then
        Nz = fallback
    ElseIf IsEmpty(v) Or IsNull(v) Then
        Nz = fallback
    Else
        Nz = CStr(v)
    End If
End Function

Private Function FirstLine(ByVal s As String) As String
    Dim lines() As String
    s = Replace(s, vbCrLf, vbLf)
    lines = Split(s, vbLf)

    Dim i As Long
    For i = LBound(lines) To UBound(lines)
        Dim t As String: t = Trim(lines(i))
        If Len(t) > 0 Then
            FirstLine = t
            Exit Function
        End If
    Next
    FirstLine = ""
End Function

Private Function ExtractSummary(ByVal description As String) As String
    ' [要約]タグの抽出を試みる
    Dim startTag As String
    Dim endTag As String
    startTag = "[要約]"
    endTag = "[/要約]"

    Dim startPos As Long
    startPos = InStr(1, description, startTag, vbTextCompare)

    If startPos > 0 Then
        Dim endPos As Long
        endPos = InStr(startPos + Len(startTag), description, endTag, vbTextCompare)

        If endPos > startPos Then
            Dim summary As String
            summary = Mid(description, startPos + Len(startTag), endPos - startPos - Len(startTag))
            ExtractSummary = Trim(summary)
            Exit Function
        End If
    End If

    ' タグがない場合は従来のFirstLine処理
    ExtractSummary = FirstLine(description)
End Function

Private Function CleanTitle(ByVal subject As String) As String
    Dim patterns As Variant
    patterns = GetConfigPatterns("TitleCleaning", "Pattern")

    If IsEmpty(patterns) Then
        CleanTitle = subject
        Exit Function
    End If

    Dim result As String
    result = subject

    Dim pattern As Variant
    For Each pattern In patterns
        If Len(Trim(pattern)) > 0 Then
            On Error Resume Next ' 不正な正規表現はスキップ

            Dim regex As Object
            Set regex = CreateObject("VBScript.RegExp")
            regex.pattern = pattern
            regex.Global = True
            regex.IgnoreCase = False

            result = regex.Replace(result, "")

            On Error GoTo 0
        End If
    Next

    CleanTitle = result
End Function

Private Function YmdOrBlank(ByVal it As Object, ByVal key As String) As String
    If HasKey(it, key) Then
        YmdOrBlank = Replace(CStr(it(key)), "-", "/")
    Else
        YmdOrBlank = "----/--/--"
    End If
End Function

Private Function AssigneeOrUnassigned(ByVal it As Object) As String
    If HasKey(it, "assigned_to") Then
        AssigneeOrUnassigned = Nz(it("assigned_to")("name"), "担当者未定")
    Else
        AssigneeOrUnassigned = "担当者未定"
    End If
End Function

Private Sub OutputToTable(ByVal byId As Object, ByVal childrenByParent As Object, ByVal roots As Collection)
    ' 出力行数をカウント
    Dim rowCount As Long
    rowCount = 1 ' ヘッダー行

    Dim p As Variant
    For Each p In roots
        If childrenByParent.Exists(CLng(p("id"))) Then
            rowCount = rowCount + childrenByParent(CLng(p("id"))).Count
        End If
    Next

    If rowCount = 1 Then
        MsgBox "出力するチケットがありません。", vbInformation
        Exit Sub
    End If

    ' 配列を準備
    Dim outputData() As Variant
    ReDim outputData(1 To rowCount, 1 To 7)

    ' ヘッダー
    outputData(1, 1) = "親タスク"
    outputData(1, 2) = "タスク名"
    outputData(1, 3) = "ステータス"
    outputData(1, 4) = "開始日"
    outputData(1, 5) = "終了日"
    outputData(1, 6) = "担当者"
    outputData(1, 7) = "要約"

    ' データ行
    Dim currentRow As Long
    currentRow = 2

    For Each p In roots
        Dim parentTitle As String
        parentTitle = CleanTitle(Nz(p("subject"), ""))

        If childrenByParent.Exists(CLng(p("id"))) Then
            Dim c As Variant
            For Each c In childrenByParent(CLng(p("id")))
                outputData(currentRow, 1) = parentTitle
                outputData(currentRow, 2) = CleanTitle(Nz(c("subject"), ""))
                outputData(currentRow, 3) = Nz(c("status")("name"), "")
                outputData(currentRow, 4) = YmdOrBlank(c, "start_date")
                outputData(currentRow, 5) = YmdOrBlank(c, "due_date")
                outputData(currentRow, 6) = AssigneeOrUnassigned(c)
                outputData(currentRow, 7) = ExtractSummary(Nz(c("description"), ""))

                currentRow = currentRow + 1
            Next
        End If
    Next

    ' シートに出力
    Dim ws As Worksheet
    Set ws = ThisWorkbook.Worksheets(1)
    ws.Cells.Clear

    Dim outputRange As Range
    Set outputRange = ws.Range(ws.Cells(1, 1), ws.Cells(rowCount, 7))
    outputRange.Value = outputData

    ' テーブル化
    Dim tbl As ListObject
    Set tbl = ws.ListObjects.Add(xlSrcRange, outputRange, , xlYes)
    tbl.Name = "RedmineIssues"
    tbl.TableStyle = "TableStyleMedium2"

    ' 書式設定
    With ws
        .Rows(1).Font.Bold = True
        .Columns("D:E").NumberFormat = "yyyy/mm/dd"
        .Columns.AutoFit
    End With

    MsgBox "出力完了: " & (rowCount - 1) & " 件のチケット", vbInformation
End Sub

' ===== 設定ファイル読み込み =====
Private Config As Object

Private Function GetConfigPath() As String
    GetConfigPath = ThisWorkbook.Path & "\redmine.config"
End Function

Private Sub LoadConfig()
    Set Config = CreateObject("Scripting.Dictionary")

    Dim fso As Object
    Set fso = CreateObject("Scripting.FileSystemObject")

    Dim configPath As String
    configPath = GetConfigPath()

    If Not fso.FileExists(configPath) Then
        ' 設定ファイルが存在しない場合は空のまま（デフォルト値を使用）
        Exit Sub
    End If

    Dim file As Object
    Set file = fso.OpenTextFile(configPath, 1) ' 1 = ForReading

    Dim currentSection As String
    currentSection = ""

    Do While Not file.AtEndOfStream
        Dim line As String
        line = Trim(file.ReadLine)

        ' 空行とコメント行をスキップ
        If Len(line) = 0 Or Left(line, 1) = ";" Then
            GoTo NextLine
        End If

        ' セクションヘッダー [Section]
        If Left(line, 1) = "[" And Right(line, 1) = "]" Then
            currentSection = Mid(line, 2, Len(line) - 2)
            GoTo NextLine
        End If

        ' key=value の解析
        Dim eqPos As Long
        eqPos = InStr(line, "=")
        If eqPos > 0 Then
            Dim key As String
            Dim value As String
            key = Trim(Left(line, eqPos - 1))
            value = Trim(Mid(line, eqPos + 1))

            If Len(currentSection) > 0 And Len(key) > 0 Then
                Config(currentSection & "." & key) = value
            End If
        End If

NextLine:
    Loop

    file.Close
End Sub

Private Function GetConfigValue(ByVal section As String, ByVal key As String, ByVal defaultValue As String) As String
    If Config Is Nothing Then LoadConfig

    Dim fullKey As String
    fullKey = section & "." & key

    If Config.Exists(fullKey) Then
        GetConfigValue = Config(fullKey)
    Else
        GetConfigValue = defaultValue
    End If
End Function

Private Function GetConfigPatterns(ByVal section As String, ByVal prefix As String) As Variant
    If Config Is Nothing Then LoadConfig

    Dim patterns As New Collection
    Dim i As Long
    i = 1

    Do
        Dim fullKey As String
        fullKey = section & "." & prefix & i

        If Config.Exists(fullKey) Then
            patterns.Add Config(fullKey)
            i = i + 1
        Else
            Exit Do
        End If
    Loop

    ' Collectionを配列に変換
    If patterns.Count > 0 Then
        Dim arr() As Variant
        ReDim arr(1 To patterns.Count)

        Dim j As Long
        For j = 1 To patterns.Count
            arr(j) = patterns(j)
        Next

        GetConfigPatterns = arr
    Else
        GetConfigPatterns = Empty
    End If
End Function