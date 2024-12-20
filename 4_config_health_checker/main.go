package main

import (
    "encoding/json"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "time"
)

type Provider struct {
    URL        string `json:"url"`
    AuthHeader string `json:"auth_header"`
}

type CheckerConfig struct {
    IntervalSeconds int `json:"interval_seconds"`
}

func main() {
    // Чтение checker_config.json
    configData, err := ioutil.ReadFile("checker_config.json")
    if err != nil {
        log.Fatalf("failed to read checker_config.json: %v", err)
    }

    var config CheckerConfig
    if err := json.Unmarshal(configData, &config); err != nil {
        log.Fatalf("failed to unmarshal checker_config.json: %v", err)
    }

    if config.IntervalSeconds <= 0 {
        config.IntervalSeconds = 60 // дефолтное значение, если что-то не так
    }

    // Канал для остановки (на случай, если хотим корректно завершать)
    stopChan := make(chan struct{})

    // Периодическая задача
    go func() {
        ticker := time.NewTicker(time.Duration(config.IntervalSeconds) * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                // Каждые interval_seconds читаем default_providers.json и пишем первые две записи в providers.json
                err := updateProviders()
                if err != nil {
                    log.Printf("error updating providers: %v", err)
                    // Приложение не падает, но логируем ошибку
                    // Если надо упасть, можем вызвать os.Exit(1), но лучше оставаться живым.
                }
            case <-stopChan:
                return
            }
        }
    }()

    // HTTP хендлеры
    http.HandleFunc("/providers", func(w http.ResponseWriter, r *http.Request) {
        // Отдаём содержимое providers.json
        f, err := os.Open("providers.json")
        if err != nil {
            http.Error(w, "failed to open providers.json", http.StatusInternalServerError)
            return
        }
        defer f.Close()

        w.Header().Set("Content-Type", "application/json")
        if _, err := io.Copy(w, f); err != nil {
            http.Error(w, "failed to read providers.json", http.StatusInternalServerError)
            return
        }
    })

    // health-check endpoint
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    })

    // Инициализируем providers.json при старте
    if err := updateProviders(); err != nil {
        log.Printf("initial update providers failed: %v", err)
    }

    // Запуск сервера
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("starting server on :%s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}

func updateProviders() error {
    defaultData, err := ioutil.ReadFile("default_providers.json")
    if err != nil {
        return err
    }

    var providers []Provider
    if err := json.Unmarshal(defaultData, &providers); err != nil {
        return err
    }

    if len(providers) < 2 {
        return nil // или ошибка, если принципиально нужно 2 записи
    }

    // Берём первые две записи
    selected := providers[:2]

    outData, err := json.MarshalIndent(selected, "", "  ")
    if err != nil {
        return err
    }

    return ioutil.WriteFile("providers.json", outData, 0644)
}
