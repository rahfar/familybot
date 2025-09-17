package db

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps redis.Client and provides a high-level interface for caching operations
type Client struct {
	client *redis.Client
}

// NewClient creates a new Redis client with the given address
func NewClient(addr string) *Client {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &Client{client: rdb}
}

// Get retrieves a value from Redis by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set stores a value in Redis with the given key and TTL
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.SetArgs(ctx, key, value, redis.SetArgs{TTL: ttl}).Err()
}

// Exists checks if a key exists in Redis
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	return result > 0, err
}

// Delete removes a key from Redis
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// Ping checks if Redis is responding
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// GetRawClient returns the underlying redis.Client for advanced operations
// This should be used sparingly and only when the high-level interface is insufficient
func (c *Client) GetRawClient() *redis.Client {
	return c.client
}

// Key generation functions for different services

// CurrencyKey generates a cache key for currency API data
func (c *Client) CurrencyKey(date time.Time) string {
	return "currencyapi_" + date.Format("2006-01-02")
}

// WeatherKey generates a cache key for weather API data
func (c *Client) WeatherKey(lat, lon float64) string {
	return fmt.Sprintf("openweatherapi_lat=%f&lon=%f", lat, lon)
}

// DeepLKey generates a cache key for DeepL translation API data
func (c *Client) DeepLKey(text []string) string {
	concatenatedString := strings.Join(text, "")
	hashBytes := md5.Sum([]byte(concatenatedString))
	hashSlice := hashBytes[:]
	return "deeplapi_" + hex.EncodeToString(hashSlice)
}

// Domain-specific cache operations

// GetCurrencyRates retrieves cached currency rates for a specific date
func (c *Client) GetCurrencyRates(ctx context.Context, date time.Time) (string, error) {
	return c.Get(ctx, c.CurrencyKey(date))
}

// SetCurrencyRates caches currency rates for a specific date with 7-day TTL
func (c *Client) SetCurrencyRates(ctx context.Context, date time.Time, data interface{}) error {
	return c.Set(ctx, c.CurrencyKey(date), data, 7*24*time.Hour)
}

// GetWeatherData retrieves cached weather data for specific coordinates
func (c *Client) GetWeatherData(ctx context.Context, lat, lon float64) (string, error) {
	return c.Get(ctx, c.WeatherKey(lat, lon))
}

// SetWeatherData caches weather data for specific coordinates with 3-hour TTL
func (c *Client) SetWeatherData(ctx context.Context, lat, lon float64, data interface{}) error {
	return c.Set(ctx, c.WeatherKey(lat, lon), data, 3*time.Hour)
}

// GetTranslation retrieves cached translation for the given text
func (c *Client) GetTranslation(ctx context.Context, text []string) (string, error) {
	return c.Get(ctx, c.DeepLKey(text))
}

// SetTranslation caches translation with 24-hour TTL
func (c *Client) SetTranslation(ctx context.Context, text []string, translation interface{}) error {
	return c.Set(ctx, c.DeepLKey(text), translation, 24*time.Hour)
}