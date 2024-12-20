#!/bin/bash

# Создание Docker сети (если не существует)
docker network create rpc-network || true

# Сборка образа
docker build -t config-health-checker .

# Удаление существующего контейнера (если есть)
docker rm -f config-health-checker || true

# Запуск контейнера в сети
docker run -it --name config-health-checker --network rpc-network -p 8081:8080 config-health-checker
