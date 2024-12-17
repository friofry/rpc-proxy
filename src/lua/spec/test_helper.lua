-- Мок для переменной ngx
_G.ngx = {
    log = function(...) print(...) end,  -- Заменяем ngx.log на print для вывода
    ERR = "ERROR",
    INFO = "INFO",
    DEBUG = "DEBUG",
    WARN = "WARN"
}
