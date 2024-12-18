# Создание образа
docker build -t openresty-example-3 .

# Удаление существующего контейнера (если есть)
docker rm -f openresty-example-3 || true


docker run -it --rm --name openresty-example-3 -p 8080:8080 openresty-example-3
