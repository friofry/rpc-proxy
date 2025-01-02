package rpc_provider

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// RpcProviderTestSuite определяет структуру тестового набора
type RpcProviderTestSuite struct {
	suite.Suite
	tempDir     string
	tempFile    string
	validJSON   string
	invalidJSON string
}

// SetupSuite выполняется перед запуском всех тестов в наборе
func (suite *RpcProviderTestSuite) SetupSuite() {
	// Создаём временную директорию для тестов
	dir, err := os.MkdirTemp("", "rpc_provider_test")
	if err != nil {
		suite.T().Fatalf("Failed to create temp dir: %v", err)
	}
	suite.tempDir = dir

	// Путь к временным файлам
	suite.tempFile = filepath.Join(suite.tempDir, "providers_test.json")

	// Определяем корректный JSON
	suite.validJSON = `{
  "providers": [
    {
      "name": "InfuraMainnet",
      "url": "https://mainnet.infura.io/v3",
      "enabled": true,
      "authType": "token-auth",
      "authToken": "infura-token"
    },
    {
      "name": "AlchemyMainnet",
      "url": "https://eth-mainnet.alchemyapi.io/v2",
      "enabled": true,
      "authType": "token-auth",
      "authToken": "alchemy-token"
    },
    {
      "name": "Example",
      "url": "https://another-provider.example.io/v2",
      "enabled": true,
      "authType": "no-auth"
    }
  ]
}`

	// Определяем некорректный JSON
	suite.invalidJSON = `{
  "providers": [
    {
      "name": "BadProvider",
      "url": "https://bad-provider.example.io"
      "enabled": true,
      "authType": "no-auth"
    }
  ]` // Обратите внимание на пропущенную запятую и закрывающую скобку
}

// TearDownSuite выполняется после завершения всех тестов в наборе
func (suite *RpcProviderTestSuite) TearDownSuite() {
	// Удаляем временную директорию и все её содержимое
	os.RemoveAll(suite.tempDir)
}

// SetupTest выполняется перед каждым тестом
func (suite *RpcProviderTestSuite) SetupTest() {
	// Перед каждым тестом очищаем файл, если он существует
	if _, err := os.Stat(suite.tempFile); err == nil {
		os.Remove(suite.tempFile)
	}
}

// TearDownTest выполняется после каждого теста
func (suite *RpcProviderTestSuite) TearDownTest() {
	// Можно добавить дополнительные действия после каждого теста, если необходимо
}

// TestReadRpcProvidersSuccess проверяет успешное чтение корректного JSON-файла
func (suite *RpcProviderTestSuite) TestReadRpcProvidersSuccess() {
	// Записываем корректный JSON в временный файл
	err := os.WriteFile(suite.tempFile, []byte(suite.validJSON), 0644)
	suite.Require().NoError(err, "Failed to write valid JSON to temp file")

	// Читаем провайдеров из файла
	providers, err := ReadRpcProviders(suite.tempFile)
	suite.Require().NoError(err, "ReadRpcProviders() returned an error")

	// Проверяем количество провайдеров
	suite.Equal(3, len(providers), "Expected 3 providers")

	// Проверяем поля первого провайдера
	first := providers[0]
	suite.Equal("InfuraMainnet", first.Name, "First provider name mismatch")
	suite.Equal("https://mainnet.infura.io/v3", first.URL, "First provider URL mismatch")
	suite.True(first.Enabled, "First provider should be enabled")
	suite.Equal(TokenAuth, first.AuthType, "First provider AuthType mismatch")
	suite.Equal("infura-token", first.AuthToken, "First provider AuthToken mismatch")
}

// TestReadRpcProvidersFileNotFound проверяет, что функция возвращает ошибку для несуществующего файла
func (suite *RpcProviderTestSuite) TestReadRpcProvidersFileNotFound() {
	_, err := ReadRpcProviders(filepath.Join(suite.tempDir, "non_existent.json"))
	suite.Error(err, "Expected error for non-existent file")
}

// TestReadRpcProvidersInvalidJSON проверяет, что функция возвращает ошибку для некорректного JSON
func (suite *RpcProviderTestSuite) TestReadRpcProvidersInvalidJSON() {
	// Записываем некорректный JSON в временный файл
	err := ioutil.WriteFile(suite.tempFile, []byte(suite.invalidJSON), 0644)
	suite.Require().NoError(err, "Failed to write invalid JSON to temp file")

	// Пытаемся прочитать провайдеров из файла
	_, err = ReadRpcProviders(suite.tempFile)
	suite.Error(err, "Expected JSON parse error")
}

// TestWriteRpcProvidersAndReadBack проверяет, что запись и последующее чтение провайдеров работают корректно
func (suite *RpcProviderTestSuite) TestWriteRpcProvidersAndReadBack() {
	// Создаём тестовых провайдеров
	wantProviders := []RpcProvider{
		{
			Name:      "TestProvider1",
			URL:       "https://test1.example.com",
			Enabled:   true,
			AuthType:  NoAuth,
			AuthToken: "",
		},
		{
			Name:      "TestProvider2",
			URL:       "https://test2.example.com",
			Enabled:   false,
			AuthType:  TokenAuth,
			AuthToken: "dummy_token",
		},
	}

	// Пишем провайдеров в файл
	err := WriteRpcProviders(suite.tempFile, wantProviders)
	suite.Require().NoError(err, "WriteRpcProviders() returned an error")

	// Читаем провайдеров из файла
	gotProviders, err := ReadRpcProviders(suite.tempFile)
	suite.Require().NoError(err, "ReadRpcProviders() returned an error")

	// Используем assert для сравнения
	assert.Equal(suite.T(), wantProviders, gotProviders, "Providers read from file do not match written providers")
}

// TestWriteRpcProvidersHandlesEmptyList проверяет, что функция корректно обрабатывает пустой список провайдеров
func (suite *RpcProviderTestSuite) TestWriteRpcProvidersHandlesEmptyList() {
	// Пишем пустой список провайдеров в файл
	err := WriteRpcProviders(suite.tempFile, []RpcProvider{})
	suite.Require().NoError(err, "WriteRpcProviders() returned an error for empty list")

	// Читаем провайдеров из файла
	gotProviders, err := ReadRpcProviders(suite.tempFile)
	suite.Require().NoError(err, "ReadRpcProviders() returned an error for empty list")

	// Проверяем, что список пуст
	suite.Empty(gotProviders, "Expected no providers in the file")
}

// TestWriteRpcProvidersInvalidPath проверяет, что функция возвращает ошибку при попытке записи в недоступный путь
func (suite *RpcProviderTestSuite) TestWriteRpcProvidersInvalidPath() {
	// Используем недопустимый путь (например, директорию вместо файла)
	err := WriteRpcProviders(suite.tempDir, []RpcProvider{})
	suite.Error(err, "Expected error when writing to a directory path")
}

// Запуск тестового набора
func TestRpcProviderTestSuite(t *testing.T) {
	suite.Run(t, new(RpcProviderTestSuite))
}
