# Docker: Golang + PostgreSQL + pgAdmin4

## docker-composeのバージョン
2.4.1

## 開発環境構築

事前にDocker実行環境を用意する。

### PostgreSQL関連の情報を記入

PostgreSQLといったサービスの設定情報を記述する。

```bash
cp docker/postgresql/.env.example docker/postgresql/.env
vi docker/postgresql/.env

cp docker/pgadmin4/.env.example docker/pgadmin4/.env
vi docker/pgadmin4/.env
```

### イメージのビルドとコンテナ起動

```bash
docker compose build # docker-compose.ymlがあるディレクトリで実行
docker compose up -d
```

イメージが未生成の場合は、時間がかかる場合があるので待つ。

### pgAdmin

http://localhost:5434

接続先のサーバーは、ホスト名 `devenv_postgresql` で登録する。

ユーザー名とパスワードは `docker/postgresql/.env` で記述した内容を入力する。
PGADMIN_DEFAULT_EMAILはメールアドレス形式の値を指定しないとエラーになり起動できない。

## Goのコードを実行する

```bash
docker exec -it devenv_app /bin/sh
```

コンテナに入ったら、コードを実行する。

```bash
go run main.go
```

## 不要になったら...

コンテナ、ネットワーク、ボリューム、イメージを削除する。

```bash
docker compose down --volumes --rmi all
```

📖 公式ドキュメント https://docs.docker.com/compose/reference/down/
