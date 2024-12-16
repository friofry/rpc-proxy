local json = require "cjson"


local function get_providers_from_cache(cache)
    -- Получение сериализованных данных из кэша
    local serialized_data = cache:get("providers_data")
    if not serialized_data then
        ngx.log(ngx.ERR, "No providers data in cache, using default")
        return {}
    end

    -- Декодирование JSON из кэша
    local ok, providers = pcall(function()
        return json.decode(serialized_data)
    end)

    -- Проверка результатов декодирования
    if not ok or not providers or type(providers) ~= "table" or #providers == 0 then
        ngx.log(ngx.ERR, "Invalid or empty providers in cache, using default")
        return {}
    end

    -- Возврат провайдеров из кэша
    return providers
end


local function cache_providers(cache, file_path)
    local file = io.open(file_path, "r")
    if not file then
        ngx.log(ngx.ERR, "Providers file not found: ", file_path)
        cache:flush_all() -- Clear any existing providers
        return
    end

    local content = file:read("*a")
    file:close()

    -- Попытка декодирования JSON
    local ok, providers = pcall(function()
        return json.decode(content)
    end)

    -- Проверка результатов декодирования
    if not ok or not providers or #providers == 0 then
        ngx.log(ngx.ERR, "Invalid or empty providers file: ", file_path, content)
        cache:flush_all() -- Очистить кэш в случае ошибки
        return
    end

    -- Сохранение провайдеров в кэше
    local serialized = json.encode(providers)
    local success, err = cache:set("providers_data", serialized)
    if not success then
        ngx.log(ngx.ERR, "Failed to save providers data in cache: ", err)
    end
end


local function get_provider_for_attempt(cache, attempt, default_providers)
    -- Получаем список провайдеров из кэша или используем провайдеры по умолчанию
    local providers = get_providers_from_cache(cache)

    -- Если попыток больше, чем количество провайдеров, возвращаем ошибку
    if attempt > #providers then
        return nil, "No more providers to try"
    end

    -- Выбираем провайдера для текущей попытки
    local provider = providers[attempt]

    -- Проверяем наличие URL провайдера
    if not provider.url or provider.url == "" then
        return nil, "Provider URL is missing"
    end

    -- Извлекаем host из URL
    local host = provider.url:gsub("^https://", "") -- Убираем https://

    return {
        host = host,
        port = 443, -- Всегда HTTPS
        auth_header = provider.auth_header -- Заголовок Authorization не используется
    }, nil
end

return {
    cache_providers = cache_providers,
    get_providers_from_cache = get_providers_from_cache,
    get_provider_for_attempt = get_provider_for_attempt
}
