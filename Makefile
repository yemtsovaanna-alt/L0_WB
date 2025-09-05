## Simple developer helpers (Kafka + Postgres)

.PHONY: migrate
migrate:
	goose -dir="./migrations" postgres "host=localhost user=postgres dbname=l0_wb password=postgres port=5432 sslmode=disable" up

.PHONY: create-kafka-topic
create-kafka-topic:
	docker exec l0wb_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --create --if-not-exists --topic orders --partitions 1 --replication-factor 1

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -a -o main ./cmd/L0_WB/

.PHONY: create-docker-image
create-docker-image:
	docker build -t l0-wb-scratch -f Dockerfile.scratch .
