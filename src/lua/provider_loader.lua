local json = require("cjson")

local M = {}

function M.reload_providers()
    ngx.log(ngx.ERR, "Reloading providers")
    local file = io.open("/usr/local/openresty/nginx/providers.json", "r")
    if file then
        local content = file:read("*all")
        file:close()
        ngx.shared.providers:set("list", content)
        ngx.log(ngx.ERR, "Providers reloaded: ", content)
    else
        ngx.log(ngx.ERR, "Failed to open providers.json")
    end
end

return M
