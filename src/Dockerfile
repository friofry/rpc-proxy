FROM openresty/openresty:alpine

# Устанавливаем perl перед использованием opm
RUN apk add --no-cache perl curl && \
    /usr/local/openresty/bin/opm install ledgetech/lua-resty-http
RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY .htpasswd /etc/nginx/.htpasswd

# Копируем конфигурационный файл Nginx в контейнер
COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf
COPY providers.json /usr/local/openresty/nginx/providers.json
COPY lua /usr/local/openresty/nginx/lua


# Указываем порт, который будет использоваться
EXPOSE 8080

# Команда запуска Nginx
CMD ["openresty", "-v"]
CMD ["openresty", "-g", "daemon off;"]
