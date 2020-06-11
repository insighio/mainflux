#!/bin/sh

OPTION_REBUILD_CONTAINERS="-r"

docker-compose -f docker/docker-compose.yml -f docker/addons/influxdb-writer/docker-compose.yml -f docker/addons/bootstrap/docker-compose.yml -f docker/addons/lora-adapter/docker-compose.yml -f docker/aedes.yml down

REBUILD_CONTAINERS=false

if [ "$1" = "$OPTION_REBUILD_CONTAINERS" ]; then
  echo "will rebuild containers."
  make users &&\
  make authn &&\
  make influxdb-writer &&\
  make lora &&\
  make docker_users &&\
  make docker_authn &&\
  make docker_influxdb-writer &&\
  make docker_lora
fi

docker network create docker_mainflux-base-net &&\
docker-compose -f docker/docker-compose.yml -f docker/addons/influxdb-writer/docker-compose.yml -f docker/addons/bootstrap/docker-compose.yml -f docker/addons/lora-adapter/docker-compose.yml -f docker/aedes.yml up -d

