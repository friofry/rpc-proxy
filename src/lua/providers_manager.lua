local resolver = require "resty.dns.resolver"
local cjson = require "cjson"

local function resolve_providers(providers)
    local resolved_providers = {}

    -- Настраиваем резолвер
    local dns, err = resolver:new{
        nameservers = {"8.8.8.8", "8.8.4.4"},
        retrans = 5,
        timeout = 2000,
    }
    if not dns then
        ngx.log(ngx.ERR, "Failed to create DNS resolver: ", err)
        return nil, err
    end

    for _, provider in ipairs(providers) do
        local host = provider.url:match("^https://([^/]+)")
        local path = provider.url:match("^https://[^/]+(/.*)$") or "/"
        local answers, err = dns:query(host)

        if not answers then
            ngx.log(ngx.ERR, "DNS query failed for ", provider.url, ": ", err)
        else
            for _, ans in ipairs(answers) do
                if ans.address then
                    table.insert(resolved_providers, {
                        ip = ans.address,
                        host = host,
                        path = path,
                        auth_header = provider.auth_header or nil
                    })
                    ngx.log(ngx.INFO, "Resolved provider: ", host, " -> ", ans.address)
                    break -- Берём первый IP
                end
            end
        end
    end

    return resolved_providers
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
        return cjson.decode(content)
    end)

    -- Проверка результатов декодирования
    if not ok or not providers or #providers == 0 then
        ngx.log(ngx.ERR, "Invalid or empty providers file: ", file_path, content)
        cache:flush_all() -- Очистить кэш в случае ошибки
        return
    end

    -- Резолв провайдеров
    local resolved_providers, err = resolve_providers(providers)
    if not resolved_providers then
        ngx.log(ngx.ERR, "Failed to resolve providers: ", err)
        return
    end

    cache:flush_all() -- Очищаем кэш перед сохранением новых записей
    for i, provider in ipairs(resolved_providers) do
        local key = tostring(i) -- Используем номер провайдера как ключ
        local value = cjson.encode({
            ip = provider.ip,
            host = provider.host,
            path = provider.path,
            auth_header = provider.auth_header
        })
        local success, err = cache:set(key, value)
        if not success then
            ngx.log(ngx.ERR, "Failed to cache provider: ", key, ", error: ", err)
        else
            ngx.log(ngx.INFO, "Cached provider [", key, "]: ", value)
        end
    end
end

local function get_provider_for_attempt(cache, attempt)
    local key = tostring(attempt) -- Формируем ключ по номеру попытки
    local cache_entry = cache:get(key)

    if not cache_entry then
        ngx.log(ngx.ERR, "Provider not found in cache for attempt: ", attempt)
        return nil, "Provider not found in cache"
    end

    local ok, data = pcall(function()
        return cjson.decode(cache_entry)
    end)

    if not ok or not data or not data.ip then
        ngx.log(ngx.ERR, "Failed to decode cache entry for attempt: ", attempt)
        return nil, "Invalid cache entry"
    end

    return {
        host = data.ip,
        port = 443,
        path = data.path,
        auth_header = data.auth_header
    }, nil
end
return {
    resolve_providers = resolve_providers,  -- Добавьте эту строку
    cache_providers = cache_providers,
    get_providers_from_cache = get_providers_from_cache,
    get_provider_for_attempt = get_provider_for_attempt
}
