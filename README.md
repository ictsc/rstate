# rstate

Terraformを利用した展開・再展開

## Usage


## Web Endpoint

## 運営むけ

> /admin

### チーム向け

> /status/:token

## APIEndPoint

> POST /admin/postJob

**Request**

```shell
curl -X POST localhost:8089/admin/postJob -d team_id=team99 -d prob_id=aks  -u user:pass
```

**Response**

追加できなかった場合、400を返す。

追加できた場合は、200を返しチームごとのWebUIのリンクを返す。


> GET /admin/list/:uuid

**Response**

```json
{
  "id": "8ed8f0d9-8eb9-11ed-9bc1-befd30ae56b9",
  "state": 0,
  "created_time": "2023-01-08T03:31:53.503152468+09:00",
  "end_time": "2023-01-08T03:32:02.914528266+09:00",
  "priority": 1,
  "team_id": "team99",
  "prob_id": "aks"
}
```

`state`は以下の通り

```go
const (
	StateUnknown   = -1
	StateSuccess   = 0
	StateRunning   = 1
	StateError     = 2
	StateWait      = 3
	StateTaskLimit = 4
)
```