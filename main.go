package main

import (
	"net/http"

	"zai-proxy/internal/config"
	"zai-proxy/internal/handler"
	"zai-proxy/internal/logger"
	"zai-proxy/internal/proxy"
	"zai-proxy/internal/version"
)

func main() {
	config.LoadConfig()
	logger.InitLogger()
	proxy.LoadProxies("proxies.txt")
	version.StartVersionUpdater()

	http.HandleFunc("/v1/models", handler.WithRequestLogging("models", handler.HandleModels))
	http.HandleFunc("/v1/chat/completions", handler.WithRequestLogging("chat", handler.HandleChatCompletions))
	http.HandleFunc("/v1/messages", handler.WithRequestLogging("messages", handler.HandleMessages))

	addr := ":" + config.Cfg.Port
	logger.LogInfo("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.LogError("Server failed: %v", err)
	}
}
