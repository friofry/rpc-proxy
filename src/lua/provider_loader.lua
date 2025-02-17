local http = require("resty.http")
local json = require("cjson")

local M = {}
local function read_config_from_url(url)
    ngx.log(ngx.ERR, "Fetching configuration from URL: ", url)
    if not url or url == "" then
        return nil, "URL is invalid or not provided"
    end

    local httpc = http.new()

    local res, err = httpc:request_uri(url, {
        method = "GET",
        headers = {
            ["Content-Type"] = "application/json",
        },
        ssl_verify = false,
    })

    if not res then
        ngx.log(ngx.ERR, "Failed to fetch configuration from URL: ", err)
        return nil, err
    end

    if res.status ~= 200 then
        ngx.log(ngx.ERR, "Non-200 response from URL: ", res.status)
        return nil, "HTTP error: " .. res.status
    end

    ngx.log(ngx.ERR, "Successfully fetched configuration from URL")
    return res.body, nil
end


-- Функция для чтения конфигурации из файла
local function read_config_from_file(filepath)
    ngx.log(ngx.ERR, "Reading configuration from file: ", filepath)
    if not filepath or filepath == "" then
        return nil, "Filepath is invalid or not provided"
    end

    local file = io.open(filepath, "r")
    if not file then
        ngx.log(ngx.ERR, "Failed to open file: ", filepath)
        return nil, "File open error"
    end

    local content = file:read("*all")
    file:close()
    ngx.log(ngx.ERR, "Successfully read configuration from file")
    return content, nil
end

-- Основная функция перезагрузки провайдеров
function M.reload_providers(premature, url, fallbackLocalConfig)
    if premature then
        return
    end

    ngx.log(ngx.ERR, "Reloading providers")

    -- Попытка загрузки конфигурации с URL
    local config, err = read_config_from_url(url)
    if not config then
        ngx.log(ngx.ERR, "Failed to load configuration from URL: ", err)

        -- Попытка загрузки конфигурации из файла
        config, err = read_config_from_file(fallbackLocalConfig)
        if not config then
            ngx.log(ngx.ERR, "Failed to load configuration from fallback file: ", err)
            return
        end
    end

    -- Parse and transform provider configuration
    local parsed_config, err = json.decode(config)
    if not parsed_config then
        ngx.log(ngx.ERR, "Failed to parse provider config: ", err)
        return
    end

    -- Clear existing providers
    ngx.shared.providers:flush_all()

    -- Store providers by chain/network
    for _, chain in ipairs(parsed_config.chains or {}) do
        local key = chain.name .. ":" .. chain.network
        ngx.shared.providers:set(key, json.encode(chain.providers))
        ngx.log(ngx.ERR, "Loaded providers for ", key, json.encode(chain.providers))
    end

    ngx.log(ngx.ERR, "Providers reloaded and stored by chain/network")
end

-- Планировщик для вызова reload_providers
function M.schedule_reload_providers(url, fallbackLocalConfig)
    local ok, err = ngx.timer.at(0, M.reload_providers, url, fallbackLocalConfig)
    if not ok then
        ngx.log(ngx.ERR, "Failed to create timer: ", err)
    end
end

return M
