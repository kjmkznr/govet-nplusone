# govet-nplusone

SQL の N+1 クエリを静的解析で検出するための go/analysis ベースのアナライザーです。まずは `database/sql` を対象に、ループ内での `Query*`/`Exec*`/`Prepare*` 呼び出しを検出してレポートします。

英語版の README は README.md を参照してください。

## 特長
- `for`/`range` などのループ内で呼ばれている `database/sql` のメソッド呼び出しを検出
- 対象メソッド: `Query`, `QueryContext`, `QueryRow`, `QueryRowContext`, `Exec`, `ExecContext`, `Prepare`, `PrepareContext`
- スタンドアロン実行（`singlechecker`）に対応。`go vet -vettool` 連携も容易

## 前提
- Go Modules 利用を前提
- Go 1.20 以降を推奨（`golang.org/x/tools` のバージョン要件に依存）

## ビルド

```shell
go build ./...
```

## インストール

```shell
go install ./cmd/nplusone
```

`$GOPATH/bin`（あるいは `GOBIN`）に `nplusone` コマンドが生成されます。

## 使い方
### スタンドアロンで解析
```
# カレントディレクトリ配下を解析
nplusone ./...

# 対象パッケージを指定
nplusone ./pkg/...
```

### go vet と連携（-vettool）
```
# 先に nplusone を go install 済みであること
VETTOOL=$(which nplusone)

go vet -vettool="$VETTOOL" ./...
```

## 出力例
次のような処理に対して:

```go
for _, id := range ids {
    _ = id
    _ = db.QueryRowContext(ctx, "SELECT 1") // want "potential N\\+1: database/sql method QueryRowContext called inside a loop"
}
```

この場合、以下のようなレポートが出力されます:

```
path/to/file.go:NN:NN: potential N+1: database/sql method QueryRowContext called inside a loop
```

## 検出ロジックの概要
- AST を `inspect` パスで走査し、`ForStmt`/`RangeStmt` に入っているかどうかを深さカウント
- ループ内の `CallExpr` だけを対象にし、型情報（`pass.TypesInfo.Selections`）からメソッドの所属パッケージが `database/sql` であるかを判定
- パッケージ関数（例: `sql.Open`）は対象外（セレクション情報が無い）

## 制限事項 / 既知の課題
- 現時点では `database/sql` のみ対象。ORM（gorm, sqlx など）やラップ関数は未対応
- N+1 の「確定」ではなく「可能性」の警告（誤検出/過検出を避けるためシンプルな規則）
- ループ外での準備（プリペアドステートメントの再利用やバッチ化）の検知・抑制は未実装
- 抑制コメントや設定ファイルによる除外機能は未実装

## ライセンス
- MIT
