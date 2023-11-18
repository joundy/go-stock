
<!--
@startuml stock

Client -> "Service Feeder"
"Service Feeder" -> Kafka Producer
"Kafka Consumer(Listen the data)" -> "Service Update Summary"
"Service Update Summary" -> Redis
	
@enduml
-->
## Intro
The service begins with the ``feeder``, which is an API used to feed subset data from the given folder `./subsetdata`. This API is used to to export subsets into the Kafka broker. The API reads all the files and pushes them one by one in order. The stock name is used as the Kafka message key to differentiate it from others. It also ensures that the same key resides in the same partition to prevent race conditions when using multiple partitions.

On the other hand, the consumer will listen for messages from the broker. It will retrieve all Kafka partitions and consume messages from each of them. The next step is updating the stock summary. The consumer will execute the ``update_summary`` method within the service. This method updates the summary based on the previous summary (if any) and the current ``order_tx`` (received from the feeder). The updated summary will be saved in Redis and grouped by ``stock_summary:<stock_name>:<day_date>``.

## How to Run

### Requirement
- Golang Version 1.20(Recommended)
- Redis Server
- Kafka/Redpanda Server
- Docker & Docker compose  to spawn kafka & redis container (not required if those running locally)
- Mockery (if you want to regenerate the mock files) https://github.com/vektra/mockery 
- Proto & Proto gen etc. (if you want to regenerate proto files) https://github.com/grpc-ecosystem/grpc-gateway\

### Local
- ``go mod download``
- ``make run-docker-panda-redis``
- ``make run``

#### Full Docker (Recommended)
- ``make run-docker-full``

### Defaults
```
# local default env for the host, change it on the Makefile & docker-compose.yaml
GRPC_ADDRESS=127.0.0.1:9090
HTTP_ADDRESS=127.0.0.1:3001
REDPANDA_ADDRESS=127.0.0.1:9092
REDIS_ADDRESS=127.0.0.1:6379
```

## API
This service will serve 2 Protocol ``HTTP & GRPC``, default is running on port 3001 for HTTP and 9090 for GRPC. 
#### GRPC
The proto file can be found in the folder ./proto ``./proto/ohlc.proto``
#### REST
```
# Feed the data from the ./subsets folder, this will be executed in the backgroud
POST http://localhost:3001/v1/feed-data

# Get feeder status, feeding data takes a while, use this api to get the status, avilable status ("IDLE" | "FEEDING" | "COMPLETED")
GET http://localhost:3001/v1/feeder-status

# Get the list of available stock summary
GET http://localhost:3001/v1/stock-list

# Get the stock summary by given stock_name & day_date: example:
GET http://localhost:3001/v1/stock-summary?stock=GGRM&day_date=2022-11-10
```
#### Example: 
Feeding data
```
POST http://localhost:3001/v1/feed-data
Command: curl -sSL --compressed -X 'POST' 'http://localhost:3001/v1/feed-data'

HTTP/1.1 200 OK
Content-Type: application/json
Grpc-Metadata-Content-Type: application/grpc
Date: Tue, 19 Sep 2023 08:42:01 GMT
Content-Length: 2

#+RESPONSE
{}
#+END
```
Feeder Status
```
GET http://localhost:3001/v1/feeder-status
Command: curl -sSL --compressed -X 'GET' 'http://localhost:3001/v1/feeder-status'
#+END
HTTP/1.1 200 OK
Content-Type: application/json
Grpc-Metadata-Content-Type: application/grpc
Date: Tue, 19 Sep 2023 08:42:20 GMT
Content-Length: 20

#+RESPONSE
{
  "status": "FEEDING"
}
#+END
```
Available Stock Summary
```
GET http://localhost:3001/v1/stock-list
Command: curl -sSL --compressed -X 'GET' 'http://localhost:3001/v1/stock-list'
#+END
HTTP/1.1 200 OK
Content-Type: application/json
Grpc-Metadata-Content-Type: application/grpc
Date: Tue, 19 Sep 2023 08:42:33 GMT
Content-Length: 345

#+RESPONSE
{
  "list": [
    {
      "stock": "UNVR",
      "dayDate": "2022-11-10"
    },
    {
      "stock": "TLKM",
      "dayDate": "2022-11-10"
    },
    {
      "stock": "HMSP",
      "dayDate": "2022-11-10"
    },
    {
      "stock": "ICBP",
      "dayDate": "2022-11-10"
    },
    {
      "stock": "BBCA",
      "dayDate": "2022-11-10"
    },
    {
      "stock": "GGRM",
      "dayDate": "2022-11-10"
    },
    {
      "stock": "BBRI",
      "dayDate": "2022-11-10"
    },
    {
      "stock": "ASII",
      "dayDate": "2022-11-10"
    }
  ]
}
#+END
```
Stock Summary
```
GET http://localhost:3001/v1/stock-summary?stock=GGRM&day_date=2022-11-10
Command: curl -sSL --compressed -X 'GET' 'http://localhost:3001/v1/stock-summary?stock=GGRM&day_date=2022-11-10'
#+END
HTTP/1.1 200 OK
Content-Type: application/json
Grpc-Metadata-Content-Type: application/grpc
Date: Tue, 19 Sep 2023 08:42:46 GMT
Content-Length: 177

#+RESPONSE
{
  "previousPrice": "22600",
  "openPrice": "22475",
  "highestPrice": "22475",
  "lowestPrice": "22100",
  "closePrice": "22200",
  "averagePrice": "22202",
  "volume": "5427",
  "value": "120492075"
}
#+END
```
