#!/bin/sh

cd ..

docker-compose -f docker/docker-compose.yml -f docker/addons/influxdb-writer/docker-compose.yml -f docker/addons/bootstrap/docker-compose.yml -f docker/addons/lora-adapter/docker-compose.yml -f docker/aedes.yml down &&\
make users &&\
make authn &&\
make influxdb-writer &&\
make docker_users &&\
make docker_authn &&\
make docker_influxdb-writer &&\
docker network create docker_mainflux-base-net &&\
docker-compose -f docker/docker-compose.yml -f docker/addons/influxdb-writer/docker-compose.yml -f docker/addons/bootstrap/docker-compose.yml -f docker/addons/lora-adapter/docker-compose.yml -f docker/aedes.yml up -d

