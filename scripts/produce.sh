#!/usr/bin/env bash
cat ./example.json | docker exec -i l0wb_kafka /opt/bitnami/kafka/bin/kafka-console-producer.sh \
  --bootstrap-server localhost:9092 \
  --topic orders
