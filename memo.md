# memo

```sh
curl -X POST localhost:8080 -d '{"record":{"value":"TGV0J3MgR28gIzEK"}}'
curl -X POST localhost:8080 -d '{"record":{"value":"TGV0J3MgR28gIzIK"}}'
curl -X POST localhost:8080 -d '{"record":{"value":"TGV0J3MgR28gIzMK"}}'
curl -X GET localhost:8080 -d '{"offset": 0}'
curl -X GET localhost:8080 -d '{"offset": 1}'
curl -X GET localhost:8080 -d '{"offset": 2}'
```

## ファイル操作

[Golang でファイルの読み込みを行う方法３選！](https://asapoon.com/article/golang-post/4869/golang-reading-file-example)

- ioutil, bufio, os

### os.File

[Go 言語の os パッケージにある File 型を使ってみる (2) ： os.File のメソッド](https://waman.hatenablog.com/entry/2017/10/04/070228)

Read：

- `n, err := f.Read(buf)`：引数に渡した容量だけ読み込む
- `n, err := f.ReadAt(buf, 9)` 第 2 引数の位置から、容量分だけ読み込む

Truncate:

- ファイルの内容を指定サイズに切り詰める

### gommap.MMap

メモリマップドファイル（ファイルマッピング）とは、ファイルをメモリ内に読み込んで、アプリケーションのアドレス空間の連続するブロックとしてファイルを操作する機能のことです。 この機能を使うと、ファイルの読み書きは適切なメモリ位置にアクセスするだけで済む

## テスト

- require.NoError()

require ：途中で Fail したらそこでテスト関数を抜ける → 前提の確認に使うのが良さそう

assert ：途中で Fail してもテストは続行する → 目的のテストを行うのが良さそう

ref: [Go で Testify でテストする際の assert と require の違い](https://qiita.com/ysti/items/a987c627d7a5e5cf32ec)

- t.Helper()

ヘルパー関数としてマークできる

ref: [Go のテストでヘルパー関数に t.Helper() を忘れない](https://qiita.com/ichiban@github/items/b5f8e5c7e00c85cb5ca7)

## e2e テスト

- httpexpect を使う： [Go でサーバのエンドツーエンドテストを行う方法](https://note.com/navitime_tech/n/ne935de0d34c9)
  - https://github.com/gavv/httpexpect
- 自力で書く： [Go でサーバを立ち上げて E2E テストを実施する CI 用のテストコードを書く](https://budougumi0617.github.io/2020/03/27/http-test-in-go/)

```go
type e2eTestSuite struct {
	suite.Suite
	srv *http.Server
}

func (s *e2eTestSuite) SetupSuite() {
	s.srv = server.NewHttpServer(":8080")
}

func (s *e2eTestSuite) SetupTest() {
	// 動的にポートを選択するので並行テストが可能。
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		fmt.Println("server run")
		if err := s.srv.Serve(l); err != http.ErrServerClosed {
			s.T().Fatalf("HTTP server ListenAndServe: %v", err)
		}
		// サーバが終了したことを通知。
		close(idleConnsClosed)
	}()
}

func (s *e2eTestSuite) TearDownSuite() {
	fmt.Println("server shutdown")
	if err := s.srv.Shutdown(context.Background()); err != nil {
		s.T().Fatalf("HTTP server Shutdown: %v", err)
	}
}

func (s *e2eTestSuite) TestXXX() {
  // api 実行 & 検証
}
```

## protobuf (プロトコルバッファ)

特徴（protobuf を使う理由）

- 一貫性のあるスキーマ

  - 例：structs と呼ぶリポジトリに protobuf とそのコンパイル済みコードを格納し、全サービスがそれに依存するようにすることで一貫性を保証

- バージョン管理

  - バージョン検査の必要性なし
  - 新機能や変更を行う際の後方互換性を保証（新フィールドの追加容易、削除も可能）
  - 削除されたフィールドを予約済み（reserved）としてマークすれば、そのフィールドを使えないようにもできる（使おうとしてもコンパイルエラーになる）

- ボイラープレートコードの削除

  - protobuf がエンコード、デコードを行うため、そのためのコードを手書きする必要がない

- 拡張性

  - protobuf コンパイラに、独自のロジックを挿入できる拡張機能をサポート
  - 例：いくつかの構造体に共通メソッドを持たせたい時に自動的に生成するプラグインを書ける

- 言語寛容性

  - 異なる言語で書かれたサービス間の通信に余計な手間をかける必要がない

- パフォーマンス
  - パフォーマンスが高い
  - データ量が小さい
  - JSON に比べて最大 6 倍の速さでシリアライズできる

### インストール

参考

- [Protocol Buffers: ざっくりとした入門](https://qiita.com/nozmiz/items/fdbd052c19dad28ab067)
- https://developers.google.com/protocol-buffers/docs/reference/go-generated

### 自動生成コマンド

```
protoc api/v1/*.proto --go_out=. --go_opt=paths=source_relative --proto_path=.
```
