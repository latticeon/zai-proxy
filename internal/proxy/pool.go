package proxy

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/proxy"

	"zai-proxy/internal/logger"
)

var (
	proxies []string
	mu      sync.RWMutex
)

func HasAvailableProxies() bool {
	mu.RLock()
	defer mu.RUnlock()
	return len(proxies) > 0
}

func shouldUseConfiguredProxy(useProxy bool) bool {
	return useProxy && HasAvailableProxies()
}

// LoadProxies 从 proxies.txt 文件加载代理列表
// 支持格式:
//   - socks5://user:pass@host:port
//   - socks5://host:port
//   - http://user:pass@host:port
//   - http://host:port
//   - https://user:pass@host:port
//   - https://host:port
//   - host:port:username:password (兼容旧 SOCKS5 格式)
//   - host:port (兼容旧 SOCKS5 格式)
func LoadProxies(path string) {
	file, err := os.Open(path)
	if err != nil {
		logger.LogInfo("No proxies.txt found, running without proxy")
		return
	}
	defer file.Close()

	var loaded []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		loaded = append(loaded, line)
	}

	mu.Lock()
	proxies = loaded
	mu.Unlock()

	logger.LogInfo("Loaded %d proxies from %s", len(loaded), path)
}

type proxyInfo struct {
	scheme   string
	addr     string
	username string
	password string
}

// getRandomProxyInfo 随机返回一个代理信息
func getRandomProxyInfo() *proxyInfo {
	mu.RLock()
	defer mu.RUnlock()

	if len(proxies) == 0 {
		return nil
	}

	line := proxies[rand.Intn(len(proxies))]
	return parseProxyLine(line)
}

func parseProxyLine(line string) *proxyInfo {
	if strings.Contains(line, "://") {
		u, err := url.Parse(line)
		if err != nil {
			logger.LogWarn("Invalid proxy URL: %s", line)
			return nil
		}
		if u.Host == "" {
			logger.LogWarn("Invalid proxy URL host: %s", line)
			return nil
		}

		scheme := strings.ToLower(u.Scheme)
		switch scheme {
		case "socks5", "http", "https":
		default:
			logger.LogWarn("Unsupported proxy scheme: %s", line)
			return nil
		}

		info := &proxyInfo{
			scheme: scheme,
			addr:   u.Host,
		}
		if u.User != nil {
			info.username = u.User.Username()
			info.password, _ = u.User.Password()
		}
		return info
	}

	parts := strings.Split(line, ":")
	switch len(parts) {
	case 2:
		return &proxyInfo{scheme: "socks5", addr: fmt.Sprintf("%s:%s", parts[0], parts[1])}
	case 4:
		return &proxyInfo{
			scheme:   "socks5",
			addr:     fmt.Sprintf("%s:%s", parts[0], parts[1]),
			username: parts[2],
			password: parts[3],
		}
	default:
		logger.LogWarn("Invalid proxy format: %s", line)
		return nil
	}
}

// GetHTTPClient 返回一个按请求配置的 http.Client。
// 仅当 useProxy=true 且存在可用代理时才会走代理，否则直连。
func GetHTTPClient(useProxy bool) *http.Client {
	if !shouldUseConfiguredProxy(useProxy) {
		return &http.Client{}
	}

	info := getRandomProxyInfo()
	if info == nil {
		return &http.Client{}
	}

	switch info.scheme {
	case "http", "https":
		proxyURL := &url.URL{
			Scheme: info.scheme,
			Host:   info.addr,
		}
		if info.username != "" {
			proxyURL.User = url.UserPassword(info.username, info.password)
		}
		logger.LogDebug("Using %s proxy: %s", strings.ToUpper(info.scheme), info.addr)
		return &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	case "", "socks5":
		logger.LogDebug("Using SOCKS5 proxy: %s", info.addr)

		var auth *proxy.Auth
		if info.username != "" {
			auth = &proxy.Auth{
				User:     info.username,
				Password: info.password,
			}
		}

		dialer, err := proxy.SOCKS5("tcp", info.addr, auth, proxy.Direct)
		if err != nil {
			logger.LogWarn("Failed to create SOCKS5 dialer: %v", err)
			return &http.Client{}
		}

		contextDialer, ok := dialer.(proxy.ContextDialer)
		if !ok {
			logger.LogWarn("SOCKS5 dialer does not support ContextDialer")
			return &http.Client{}
		}

		return &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return contextDialer.DialContext(ctx, network, addr)
				},
			},
		}
	default:
		logger.LogWarn("Unsupported proxy scheme: %s", info.scheme)
		return &http.Client{}
	}
}
