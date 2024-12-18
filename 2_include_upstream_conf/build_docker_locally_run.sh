# Создание образа
docker build -t multi-upstream .

# Удаление существующего контейнера (если есть)
docker rm -f multi-upstream || true


docker run -it --rm --name multi-upstream -p 8080:443 multi-upstream
