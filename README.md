# L0_WB
- postgres
- kafka
- docker

We could create a db and user like that
```sql
CREATE ROLE yemtsova_anna
WITH LOGIN PASSWORD 'admin!!!';

ALTER SCHEMA public OWNER TO yemtsova_anna;

GRANT CREATE, USAGE ON SCHEMA public TO yemtsova_anna;

CREATE DATABASE rwb_data
WITH OWNER = yemtsova_anna;
```
But docker image already handles that, so it's unnecessary, and we should not do that in migrations.

### Example launch
```bash
$ docker-compose up # runs kafka and postgres
$ make migrate
$ make create-kafka-topic
```

Optionally you can `make build` and `make create-docker-image` and add to compose, but just running it is easier

To publish a sample message to kafka run `/scripts/produce.sh` script.

To check if your message has been delivered open `/web/page/index.html` and hit "get message button"


## Launch

### kafka and postgres
```bash
$ docker-compose up
```

### Service
- set the following environmental variables:
```bash
DB_USER
DB_NAME
DB_PASSWORD
DB_HOST
DB_PORT
KAFKA_HOST
KAFKA_PORT
```

- run `main.go`

## Scripts

- `create-broker.sh` creates kafka broker  with `orders` topic
- `produce.sh` sends a sample message to the `orders` topic

## Migrations

Run
```
goose -dir=*migrations directory* *driver* *connection string* up`
```
Example
```
goose -dir="./migrations" postgres "user=postgres dbname=l0_wb password=postgres port=5432 sslmode=disable" up
```