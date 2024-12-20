# Создание Docker сети (если не существует)
docker network create rpc-network || true

# Создание образа
docker build -t rpc-proxy .

# Удаление существующего контейнера (если есть)
docker rm -f rpc-proxy || true

# Запуск контейнера
docker run -it --rm --name rpc-proxy --network rpc-network -p 8080:8080 rpc-proxy
