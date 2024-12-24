# Создание Docker сети (если не существует)
docker network create rpc-network || true

# Создание образа
docker build -t rpc-proxy .

# Удаление существующего контейнера (если есть)
docker rm -f rpc-proxy || true

# Запуск контейнера
docker run -it --rm \
  --name rpc-proxy \
  --network rpc-network \
  -p 8080:8080 \
  -e CONFIG_HEALTH_CHECKER_URL=http://config-health-checker:8080/providers \
  rpc-proxy


#!/bin/bash

# Создание Docker сети (если не существует)
docker network create rpc-network || true

# Сборка образа
docker build -t rpc-proxy ./rpc-proxy

# Удаление существующего контейнера (если есть)
docker rm -f rpc-proxy || true

# Запуск контейнера с переменной окружения
docker run -d --name rpc-proxy \
  --network rpc-network \
  -p 8080:8080 \

  rpc-proxy
