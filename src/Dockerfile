FROM openresty/openresty:alpine

COPY nginx.conf /etc/nginx/nginx.conf
COPY .htpasswd /etc/nginx/.htpasswd
COPY lua /etc/nginx/lua
COPY providers.json /etc/nginx/providers.json
