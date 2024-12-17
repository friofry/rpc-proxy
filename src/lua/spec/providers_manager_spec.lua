package.loaded["resty.dns.resolver"] = require("spec.mocks.resty_dns_resolver")
require("spec.test_helper")  -- Подключаем мок для ngx

local providers_manager = require("providers_manager")

describe("resolve_providers", function()
    local mock_resolver = {
        query = function(_, host)
            if host == "example.com" then
                return { { address = "127.0.0.1" } }
            elseif host == "invalid.com" then
                return nil, "DNS resolution failed"
            end
            return nil, "Unknown host"
        end,
    }

    it("should resolve valid providers", function()
        local providers = {
            { url = "https://example.com/api", auth_header = "Bearer token123" },
        }

        local resolved = providers_manager.resolve_providers(providers, mock_resolver)
        assert.is_not_nil(resolved)
        assert.are.same({
            {
                ip = "127.0.0.1",
                host = "example.com",
                path = "/api",
                auth_header = "Bearer token123",
            },
        }, resolved)
    end)

    it("should handle invalid DNS resolution", function()
        local providers = {
            { url = "https://invalid.com/api", auth_header = nil },
        }

        local resolved = providers_manager.resolve_providers(providers, mock_resolver)
        assert.is_not_nil(resolved)
        assert.are.same({}, resolved) -- Ожидается пустой результат
    end)
end)
