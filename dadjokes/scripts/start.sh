#!/bin/bash

echo "----Running go mod vendor----"
go mod download
go mod vendor
go mod tidy

echo "----Starting Docker---"
docker compose -f docker-compose.yml up --build

docker-compose down --remove-orphans
docker image prune -a -f