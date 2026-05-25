package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

func NewRateLimitMiddleware(redisClient *redis.Client, prefix string, formattedRate string) gin.HandlerFunc {
	rate, err := limiter.NewRateFromFormatted(formattedRate)
	if err != nil {
		panic(fmt.Sprintf("invalid rate limit format %q: %v", formattedRate, err))
	}

	store, err := sredis.NewStoreWithOptions(redisClient, limiter.StoreOptions{
		Prefix:   prefix,
		MaxRetry: 3,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create rate limiter store %q: %v", prefix, err))
	}

	instance := limiter.New(store, rate)
	return mgin.NewMiddleware(instance)
}
