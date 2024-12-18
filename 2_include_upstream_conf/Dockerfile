# Используем официальный базовый образ OpenResty
FROM openresty/openresty:alpine

# Копируем конфигурационный файл Nginx в контейнер
COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf
COPY upstream.conf /usr/local/openresty/nginx/upstream.conf
COPY map.conf /usr/local/openresty/nginx/map.conf
RUN mkdir -p /etc/nginx/certs
COPY certs/nginx-selfsigned.crt /etc/nginx/certs/nginx-selfsigned.crt
COPY certs/nginx-selfsigned.key /etc/nginx/certs/nginx-selfsigned.key
# Указываем порт, который будет использоваться
EXPOSE 443

# Команда запуска Nginx
CMD ["openresty", "-g", "daemon off;"]
