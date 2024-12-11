FROM nginx:alpine

# Копируем конфигурацию Nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Копируем файл с логинами и паролями
COPY .htpasswd /etc/nginx/.htpasswd
