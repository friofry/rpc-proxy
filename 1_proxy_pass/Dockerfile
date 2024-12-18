# Используем официальный базовый образ OpenResty
FROM openresty/openresty:alpine

# Копируем конфигурационный файл Nginx в контейнер
COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf

# Копируем файлы Lua (если требуется)
#COPY lua/ /usr/local/openresty/nginx/lua/

# Указываем порт, который будет использоваться
EXPOSE 80

# Команда запуска Nginx
CMD ["openresty", "-g", "daemon off;"]
