package version

import (
	"io"
	"regexp"
	"sync"
	"time"

	"zai-proxy/internal/logger"
	"zai-proxy/internal/proxy"
)

var (
	feVersion   string
	versionLock sync.RWMutex
)

func GetFeVersion() string {
	versionLock.RLock()
	defer versionLock.RUnlock()
	return feVersion
}

func fetchFeVersion() {
	resp, err := proxy.GetHTTPClient(false).Get("https://chat.z.ai/")
	if err != nil {
		logger.LogError("Failed to fetch fe version: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.LogError("Failed to read fe version response: %v", err)
		return
	}

	re := regexp.MustCompile(`prod-fe-[\.\d]+`)
	match := re.FindString(string(body))
	if match != "" {
		versionLock.Lock()
		feVersion = match
		versionLock.Unlock()
		logger.LogInfo("Updated fe version: %s", match)
	}
}

func StartVersionUpdater() {
	go func() {
		fetchFeVersion()
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			fetchFeVersion()
		}
	}()
}
