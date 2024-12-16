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
    local providers = get_providers_from_cache(cache)

    if attempt > #providers then
        return nil, "No more providers to try"
    end

    local provider = providers[attempt]
    local cache_entry = cache:get(provider.url)

    if not cache_entry then
        ngx.log(ngx.ERR, "Provider not found in cache: ", provider.url)
        return nil, "Provider not found in cache"
    end

    local ok, data = pcall(function()
        return cjson.decode(cache_entry)
    end)

    if not ok or not data or not data.ip then
        ngx.log(ngx.ERR, "Failed to decode cache entry for provider: ", provider.url)
        return nil, "Invalid cache entry"
    end

    return {
        host = data.ip,
        port = 443,
        auth_header = data.auth_header
    }, nil
end


return {
    cache_providers = cache_providers,
    get_providers_from_cache = get_providers_from_cache,
    get_provider_for_attempt = get_provider_for_attempt
}
