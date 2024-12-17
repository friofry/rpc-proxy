# Используем базовый образ Ubuntu
FROM ubuntu:22.04

# Устанавливаем необходимые зависимости
RUN apt-get update && apt-get install -y \
    curl \
    git \
    lua5.1 \
    luarocks \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем busted через luarocks
RUN luarocks install busted
RUN luarocks install lua-cjson
RUN luarocks install lua-resty-dns
RUN luarocks install luabitop

# Копируем исходный код и тесты в контейнер
WORKDIR /app
COPY . .

# Запускаем тесты при старте контейнера
CMD ["busted", "spec/"]
