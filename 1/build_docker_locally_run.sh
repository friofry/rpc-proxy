# Создание образа
docker build -t openresty-example .

# Удаление существующего контейнера (если есть)
docker rm -f openresty-example || true


docker run -it --rm --name openresty-example -p 8080:80 openresty-example
