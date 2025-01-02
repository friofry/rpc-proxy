package rpc_provider

import (
	"encoding/json"
	"os"
)

// RpcProvidersFile описывает структуру корня JSON-файла провайдеров.
type RpcProvidersFile struct {
	Providers []RpcProvider `json:"providers"`
}

// ReadRpcProviders читает список провайдеров из JSON-файла.
func ReadRpcProviders(filename string) ([]RpcProvider, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var pf RpcProvidersFile
	if err := json.NewDecoder(f).Decode(&pf); err != nil {
		return nil, err
	}
	return pf.Providers, nil
}

// WriteRpcProviders записывает список провайдеров в JSON-файл.
func WriteRpcProviders(filename string, providers []RpcProvider) error {
	pf := RpcProvidersFile{Providers: providers}

	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
