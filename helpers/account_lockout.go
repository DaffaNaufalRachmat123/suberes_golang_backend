package helpers

import (
	"context"
	"fmt"
	"time"
)

const (
	MaxLoginAttempts    = 5
	LockoutDuration     = 15 * time.Minute
	loginAttemptsKeyFmt = "login_attempts:%s:%s" // login_attempts:{user_type}:{identifier}
	lockoutKeyFmt       = "account_locked:%s:%s" // account_locked:{user_type}:{identifier}
)

// IsAccountLocked checks if the account is currently locked out.
func IsAccountLocked(userType, identifier string) bool {
	key := fmt.Sprintf(lockoutKeyFmt, userType, identifier)
	val, err := RedisClient.Get(context.Background(), key).Result()
	return err == nil && val == "1"
}

// RecordFailedLogin increments the failed attempt counter.
// Returns true if the account is now locked.
func RecordFailedLogin(userType, identifier string) bool {
	ctx := context.Background()
	attemptsKey := fmt.Sprintf(loginAttemptsKeyFmt, userType, identifier)

	count, _ := RedisClient.Incr(ctx, attemptsKey).Result()

	// Set TTL on first attempt
	if count == 1 {
		RedisClient.Expire(ctx, attemptsKey, LockoutDuration)
	}

	if count >= MaxLoginAttempts {
		// Lock the account
		lockKey := fmt.Sprintf(lockoutKeyFmt, userType, identifier)
		RedisClient.Set(ctx, lockKey, "1", LockoutDuration)
		RedisClient.Del(ctx, attemptsKey)

		WriteAuditLog(AuditLog{
			Event:    AuditAccountLocked,
			UserType: userType,
			Resource: identifier,
			Details:  fmt.Sprintf("account locked after %d failed attempts", MaxLoginAttempts),
			Success:  false,
		})

		return true
	}

	return false
}

// ClearFailedLogin resets the failed attempt counter on successful login.
func ClearFailedLogin(userType, identifier string) {
	ctx := context.Background()
	attemptsKey := fmt.Sprintf(loginAttemptsKeyFmt, userType, identifier)
	RedisClient.Del(ctx, attemptsKey)
}
