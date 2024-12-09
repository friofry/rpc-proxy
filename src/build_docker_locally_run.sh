# Создание образа
docker build -t rpc-proxy .

# Удаление существующего контейнера (если есть)
docker rm -f rpc-proxy || true

# Запуск контейнера
docker run -it --rm --name rpc-proxy -p 8080:8080 rpc-proxy
