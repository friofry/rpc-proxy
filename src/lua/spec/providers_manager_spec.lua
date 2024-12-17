package.loaded["resty.dns.resolver"] = require("spec.mocks.resty_dns_resolver")
require("spec.test_helper")  -- Подключаем мок для ngx

local providers_manager = require("providers_manager")
local cjson = require("cjson")

describe("cache_providers and get_provider_for_attempt", function()
    local mock_cache

    -- Подготовка мок-кэша
    before_each(function()
        mock_cache = {
            storage = {},
            flush_all = function(self)
                self.storage = {}
            end,
            set = function(self, key, value)
                self.storage[key] = value
                return true
            end,
            get = function(self, key)
                return self.storage[key]
            end
        }
    end)

it("should cache resolved providers successfully", function()
    -- Мок файла providers.json
    local test_file_path = "spec/test_providers.json"
    local test_content = cjson.encode({
        { url = "https://example.com/api", auth_header = "Bearer token123" },
        { url = "https://invalid.com/api", auth_header = nil }
    })

    -- Создаём тестовый файл
    local file = io.open(test_file_path, "w")
    file:write(test_content)
    file:close()

    -- Мок DNS-резолвера
    package.loaded["resty.dns.resolver"] = {
        new = function()
            return {
                query = function(_, host)
                    if host == "example.com" then
                        return { { address = "127.0.0.1" } }
                    end
                    return nil, "DNS resolution failed"
                end
            }
        end
    }

    -- Запуск функции cache_providers
    providers_manager.cache_providers(mock_cache, test_file_path)

    -- Проверка содержимого кэша
    assert.is_not_nil(mock_cache:get("1"))
    assert.is_nil(mock_cache:get("2")) -- Второй провайдер должен отсутствовать из-за ошибки DNS

    local cached_1 = cjson.decode(mock_cache:get("1"))
    assert.are.same("127.0.0.1", cached_1.ip)
    assert.are.same("example.com", cached_1.host)
    assert.are.same("/api", cached_1.path)
    assert.are.same("Bearer token123", cached_1.auth_header)
end)


    it("should retrieve provider for a given attempt", function()
        -- Добавляем данные в мок-кэш
        mock_cache:set("1", cjson.encode({
            ip = "127.0.0.1",
            host = "example.com",
            path = "/api",
            auth_header = "Bearer token123"
        }))

        -- Запуск функции get_provider_for_attempt
        local provider, err = providers_manager.get_provider_for_attempt(mock_cache, 1)

        -- Проверка результата
        assert.is_nil(err)
        assert.are.same("127.0.0.1", provider.host)
        assert.are.same(443, provider.port)
        assert.are.same("/api", provider.path)
        assert.are.same("Bearer token123", provider.auth_header)
    end)

    it("should handle missing provider in cache", function()
        -- Запуск функции get_provider_for_attempt с пустым кэшем
        local provider, err = providers_manager.get_provider_for_attempt(mock_cache, 1)

        -- Проверка результата
        assert.is_nil(provider)
        assert.are.same("Provider not found in cache", err)
    end)


    it("should return correct providers for attempts 1 to 4 and nil for attempt 5", function()
        -- Подготавливаем провайдеров
        local resolved_providers = {
            { ip = "127.0.0.1", host = "example1.com", path = "/api1", auth_header = "token1" },
            { ip = "127.0.0.2", host = "example2.com", path = "/api2", auth_header = "token2" },
            { ip = "127.0.0.3", host = "example3.com", path = "/api3", auth_header = "token3" },
            { ip = "127.0.0.4", host = "example4.com", path = "/api4", auth_header = "token4" },
        }

        -- Добавляем провайдеров в кэш
        for i, provider in ipairs(resolved_providers) do
            local key = tostring(i)
            local value = cjson.encode(provider)
            mock_cache:set(key, value)
        end

        -- Проверяем попытки с 1 по 4
        for i = 1, 4 do
            local provider, err = providers_manager.get_provider_for_attempt(mock_cache, i)
            assert.is_nil(err)
            assert.is_not_nil(provider)
            assert.are.same(resolved_providers[i].ip, provider.host)
            assert.are.same(443, provider.port)
            assert.are.same(resolved_providers[i].path, provider.path)
            assert.are.same(resolved_providers[i].auth_header, provider.auth_header)
        end

        -- Проверяем попытку 5 (должен вернуть nil и ошибку)
        local provider, err = providers_manager.get_provider_for_attempt(mock_cache, 5)
        assert.is_nil(provider)
        assert.are.same("Provider not found in cache", err)
    end)
end)
