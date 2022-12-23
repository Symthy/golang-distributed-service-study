# memo

```sh
curl -X POST localhost:8080 -d '{"record":{"value":"TGV0J3MgR28gIzEK"}}'
curl -X POST localhost:8080 -d '{"record":{"value":"TGV0J3MgR28gIzIK"}}'
curl -X POST localhost:8080 -d '{"record":{"value":"TGV0J3MgR28gIzMK"}}'
curl -X GET localhost:8080 -d '{"offset": 0}'
curl -X GET localhost:8080 -d '{"offset": 1}'
curl -X GET localhost:8080 -d '{"offset": 2}'
```

e2e

- httpexpect を使う： [Go でサーバのエンドツーエンドテストを行う方法](https://note.com/navitime_tech/n/ne935de0d34c9)
  - https://github.com/gavv/httpexpect
- 自力で書く： [Go でサーバを立ち上げて E2E テストを実施する CI 用のテストコードを書く](https://budougumi0617.github.io/2020/03/27/http-test-in-go/)
