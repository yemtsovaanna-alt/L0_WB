#!/usr/bin/env bash
docker exec l0wb_kafka /opt/bitnami/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create \
  --if-not-exists \
  --topic orders \
  --partitions 1 \
  --replication-factor 1 \
  --config retention.ms=31536000000
