local json = require "cjson"
local url_parser = require "resty.url"

local function load_providers_from_file(file_path, default_providers)
    local file = io.open(file_path, "r")
    if not file then
        ngx.log(ngx.ERR, "Failed to open providers config file: ", file_path)
        return default_providers
    end

    local content = file:read("*a")
    file:close()

    local ok, parsed
    pcall(function()
        parsed = json.decode(content)
    end)

    if not ok or not parsed or not parsed.providers or #parsed.providers == 0 then
        ngx.log(ngx.ERR, "Invalid providers config format in file: ", file_path)
        return default_providers
    end

    return parsed.providers
end

local function cache_providers(cache, file_path, default_providers)
    local providers = load_providers_from_file(file_path, default_providers)
    local serialized = json.encode(providers)
    local success, err = cache:set("providers_data", serialized)
    if not success then
        ngx.log(ngx.ERR, "Failed to save providers data in cache: ", err)
    end
end

local function get_providers_from_cache(cache, default_providers)
    local serialized_data = cache:get("providers_data")
    if not serialized_data then
        ngx.log(ngx.ERR, "No providers data in cache, using default")
        return default_providers
    end

    local ok, providers
    pcall(function()
        providers = json.decode(serialized_data)
    end)

    if not ok or not providers or #providers == 0 then
        ngx.log(ngx.ERR, "Invalid or empty providers in cache, using default")
        return default_providers
    end

    return providers
end

local function get_provider_for_attempt(cache, attempt, default_providers)
    local providers = get_providers_from_cache(cache, default_providers)

    if attempt > #providers then
        return nil, "No more providers to try"
    end

    local provider = providers[attempt]
    local parsed = url_parser.parse(provider.url)
    if not parsed then
        return nil, "Invalid provider URL: " .. provider.url
    end

    local host = parsed.host
    local port = parsed.port or (parsed.scheme == "https" and 443 or 80)

    return {
        host = host,
        port = port,
        auth_header = provider.auth_header or nil
    }, nil
end

return {
    cache_providers = cache_providers,
    get_provider_for_attempt = get_provider_for_attempt
}
