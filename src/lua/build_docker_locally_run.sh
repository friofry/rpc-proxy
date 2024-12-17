# Создание образа
docker build -t lua-test-runner .


# Удаление существующего контейнера (если есть)
docker rm -f lua-test-runner || true

# Запуск контейнера
docker run --rm lua-test-runner
