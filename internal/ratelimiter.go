package internal

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jyouturer/gmail-ai/internal/logging"
	"go.uber.org/zap"
)

type DistributedRateLimiter struct {
	client   *redis.Client
	interval time.Duration
	maxCalls int64
	requests int
}

func NewDistributedRateLimiter(client *redis.Client, interval time.Duration, maxCalls int64, requests int) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		client:   client,
		interval: interval,
		maxCalls: maxCalls,
		requests: requests,
	}
}

func (d *DistributedRateLimiter) CallAPI() {
	ctx := context.Background()

	for i := 0; i < d.requests; i++ {
		// Get the current timestamp in seconds
		now := time.Now().Unix()

		// Use a Lua script to atomically update the Redis counter
		luaScript := `
local key = KEYS[1]
local maxCalls = tonumber(ARGV[1])
local interval = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

-- Clear outdated calls
redis.call('ZREMRANGEBYSCORE', key, 0, now - interval)

-- Count the remaining calls
local currentCalls = redis.call('ZCARD', key)

if currentCalls < maxCalls then
  redis.call('ZADD', key, now, now)
  redis.call('EXPIRE', key, interval)
  return 1
else
  return 0
end
`
		// Run the Lua script and check the result
		result, err := d.client.Eval(ctx, luaScript, []string{"my_api_key"}, d.maxCalls, d.interval, now).Result()
		if err != nil {
			logging.Logger.Info("Error executing Lua script:", zap.Error(err))
			return
		}

		if result == int64(1) {
			// Make your API call here
			logging.Logger.Info("API request:", zap.Int("request", i+1))
		} else {
			logging.Logger.Info("Rate limit exceeded")
			time.Sleep(d.interval)
		}
	}
}
