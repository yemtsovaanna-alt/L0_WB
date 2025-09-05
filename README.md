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

And we also could create lots of tables like this
```sql
-- 1) Заказы
CREATE TABLE orders (
    order_uid          text PRIMARY KEY,
    track_number       text NOT NULL,
    entry              text NOT NULL,
    locale             text NOT NULL,
    internal_signature text,
    customer_id        text NOT NULL,
    delivery_service   text NOT NULL,
    shardkey           text NOT NULL,
    sm_id              integer NOT NULL,
    date_created       timestamptz NOT NULL,
    oof_shard          text NOT NULL
);

-- 2) Доставка (1:1 с заказом)
CREATE TABLE deliveries (
    order_uid    text PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name         text NOT NULL,
    phone        text NOT NULL,
    zip          text,
    city         text NOT NULL,
    address      text NOT NULL,
    region       text,
    email        text
);

-- 3) Оплата (1:1 с заказом)
CREATE TABLE payments (
    transaction   text PRIMARY KEY,
    order_uid     text UNIQUE NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    request_id    text,
    currency      text NOT NULL,
    provider      text NOT NULL,
    amount        integer NOT NULL,
    payment_dt    bigint  NOT NULL,  -- unix-epoch (секунды)
    bank          text,
    delivery_cost integer NOT NULL,
    goods_total   integer NOT NULL,
    custom_fee    integer NOT NULL
);

-- 4) Позиции заказа (1:М)
CREATE TABLE order_items (
    id            bigserial PRIMARY KEY,
    order_uid     text NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id       bigint NOT NULL,
    track_number  text NOT NULL,
    price         integer NOT NULL,
    rid           text NOT NULL,
    name          text NOT NULL,
    sale          integer NOT NULL,
    size          text NOT NULL,
    total_price   integer NOT NULL,
    nm_id         bigint NOT NULL,
    brand         text NOT NULL,
    status        integer NOT NULL
);
```
But it was just easier to use jsonb.

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