package config

import rpcprovider "github.com/friofry/config-health-checker/rpcprovider"

// Config представляет основную структуру конфигурационного файла.
type Config struct {
	IntervalSeconds   int                     `json:"interval_seconds"`
	ReferenceProvider rpcprovider.RpcProvider `json:"reference_provider"`
}
