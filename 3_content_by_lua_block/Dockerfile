# Используем официальный базовый образ OpenResty
FROM openresty/openresty:alpine


# Копируем файлы Lua (если требуется)
#COPY lua/ /usr/local/openresty/nginx/lua/
# Устанавливаем perl перед использованием opm
RUN apk add --no-cache perl curl && \
    /usr/local/openresty/bin/opm install ledgetech/lua-resty-http
RUN apk add --no-cache ca-certificates && update-ca-certificates

# Копируем конфигурационный файл Nginx в контейнер
COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf
COPY providers.json /usr/local/openresty/nginx/providers.json


# Указываем порт, который будет использоваться
EXPOSE 8080

# Команда запуска Nginx
CMD ["openresty", "-g", "daemon off;"]
