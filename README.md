# explorer-api

## Build With Source
```bash
git clone git@github.com:inscription-c/explorer-api.git
cd explorer-api
go build
```

## Run
```bash
./explorer-api -c <path_to_config>/config.yaml
```

config example
```yaml
server:
  name: "explorer"
  testnet: true
  rpc_listen: ":8336"
  pprof: false
  prometheus: false
chain:
  url: "http://127.0.0.1:18334"
  username: "root"
  password: "root"
  start_height: 0
db:
  mysql:
    addr: "127.0.0.1:3306"
    user: "root"
    password: "root"
    db: "explorer"
  indexer:
    addr: "127.0.0.1:3306"
    user: "root"
    password: "root"
    db: "cins"
sentry:
  dsn: ""
  traces_sample_rate: 1.0
origins:
  - ".*"
```