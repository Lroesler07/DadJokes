#!/bin/bash

echo "----Starting Docker----"
docker compose -f docker-compose.yml up --build

echo "----Shutting down----"
docker-compose down --remove-orphans
docker image prune -a -f