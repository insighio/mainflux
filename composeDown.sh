#!/bin/bash

MF_UI_PORT=3000 \
UI_PORT=3000 \
MF_MQTT_ADAPTER_PORT=1884 \
RPROXY_PORT=3003 \
UI_INSIGHIO_PORT=3004 \
docker-compose -f docker/docker-compose.yml -f docker/addons/influxdb-writer/docker-compose.yml -f docker/addons/bootstrap/docker-compose.yml -f docker/addons/lora-adapter/docker-compose.yml down
