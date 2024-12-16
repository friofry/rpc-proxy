FROM openresty/openresty:alpine

COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf

COPY .htpasswd /etc/nginx/.htpasswd

COPY lua /usr/local/openresty/nginx/lua
COPY providers.json /usr/local/openresty/nginx/providers.json

# Указываем порт, который будет использоваться
EXPOSE 80

# Команда запуска Nginx
CMD ["openresty", "-g", "daemon off;"]
