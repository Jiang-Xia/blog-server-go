// Package buff 天气临时 EXP 加成。
package buff

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"
)

// WeatherBuffInfo 天气 Buff 信息。
type WeatherBuffInfo struct {
	Label    string  `json:"label"`
	ExpBoost float64 `json:"expBoost"`
	Weather  string  `json:"weather"`
}

// WeatherService 天气 Buff 查询。
type WeatherService struct {
	client *http.Client
	cache  map[string]weatherCacheEntry
	mu     sync.Mutex
}

type weatherCacheEntry struct {
	at    time.Time
	value *WeatherBuffInfo
}

const weatherCacheTTL = 10 * time.Minute

// NewWeatherService 构造 WeatherService。
func NewWeatherService() *WeatherService {
	return &WeatherService{
		client: &http.Client{Timeout: 5 * time.Second},
		cache:  make(map[string]weatherCacheEntry),
	}
}

// GetWeatherBuff 根据城市返回临时 EXP 加成；失败返回 nil。
func (s *WeatherService) GetWeatherBuff(ctx context.Context, city string) (*WeatherBuffInfo, error) {
	if city == "" {
		city = "北京"
	}
	s.mu.Lock()
	if c, ok := s.cache[city]; ok && time.Since(c.at) < weatherCacheTTL {
		s.mu.Unlock()
		return c.value, nil
	}
	s.mu.Unlock()

	info := s.fetchOrStub(ctx, city)

	s.mu.Lock()
	s.cache[city] = weatherCacheEntry{at: time.Now(), value: info}
	s.mu.Unlock()
	return info, nil
}

func (s *WeatherService) fetchOrStub(ctx context.Context, city string) *WeatherBuffInfo {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://wttr.in/"+city+"?format=3", nil)
	if err != nil {
		return stubWeather(city)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return stubWeather(city)
	}
	defer resp.Body.Close()
	buf := make([]byte, 256)
	n, _ := resp.Body.Read(buf)
	text := strings.TrimSpace(string(buf[:n]))
	if text == "" {
		return stubWeather(city)
	}
	return mapWeatherToBuff(text)
}

func stubWeather(city string) *WeatherBuffInfo {
	return &WeatherBuffInfo{Label: "默认 EXP+3%", ExpBoost: 0.03, Weather: city + "（离线默认）"}
}

func mapWeatherToBuff(weather string) *WeatherBuffInfo {
	lower := strings.ToLower(weather)
	if strings.Contains(weather, "雨") || strings.Contains(lower, "rain") {
		return &WeatherBuffInfo{Label: "雨天 EXP+10%", ExpBoost: 0.1, Weather: weather}
	}
	if strings.Contains(weather, "晴") || strings.Contains(lower, "sunny") || strings.Contains(lower, "clear") {
		return &WeatherBuffInfo{Label: "晴天 EXP+5%", ExpBoost: 0.05, Weather: weather}
	}
	if strings.Contains(weather, "雪") || strings.Contains(lower, "snow") {
		return &WeatherBuffInfo{Label: "雪天 EXP+8%", ExpBoost: 0.08, Weather: weather}
	}
	return nil
}
