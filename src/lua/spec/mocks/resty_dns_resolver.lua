local resolver_mock = {}

function resolver_mock:new()
    return {
        query = function(_, host)
            if host == "example.com" then
                return { { address = "127.0.0.1" } }
            elseif host == "invalid.com" then
                return nil, "DNS resolution failed"
            end
            return nil, "Unknown host"
        end
    }
end

return resolver_mock
