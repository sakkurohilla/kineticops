/**
 * Client-side rate limiter to prevent excessive API calls
 * Uses token bucket algorithm
 */

interface RateLimiterConfig {
  maxTokens: number; // Maximum tokens (requests) allowed
  refillRate: number; // Tokens added per second
  refillInterval: number; // Milliseconds between refills
}

class RateLimiter {
  private tokens: number;
  private lastRefill: number;
  private config: RateLimiterConfig;

  constructor(config: RateLimiterConfig) {
    this.config = config;
    this.tokens = config.maxTokens;
    this.lastRefill = Date.now();
  }

  private refill(): void {
    const now = Date.now();
    const timePassed = now - this.lastRefill;
    const intervalsElapsed = Math.floor(timePassed / this.config.refillInterval);

    if (intervalsElapsed > 0) {
      const tokensToAdd = intervalsElapsed * this.config.refillRate;
      this.tokens = Math.min(this.config.maxTokens, this.tokens + tokensToAdd);
      this.lastRefill = now;
    }
  }

  canMakeRequest(): boolean {
    this.refill();
    if (this.tokens > 0) {
      this.tokens--;
      return true;
    }
    return false;
  }

  getRemainingTokens(): number {
    this.refill();
    return this.tokens;
  }

  getWaitTime(): number {
    if (this.tokens > 0) return 0;
    return this.config.refillInterval;
  }
}

// Global rate limiters for different endpoints
export const apiRateLimiter = new RateLimiter({
  maxTokens: 100, // 100 requests
  refillRate: 10, // 10 requests per second
  refillInterval: 1000, // 1 second
});

export const metricsRateLimiter = new RateLimiter({
  maxTokens: 50,
  refillRate: 5,
  refillInterval: 1000,
});

export const authRateLimiter = new RateLimiter({
  maxTokens: 5, // Strict limit for auth
  refillRate: 1,
  refillInterval: 5000, // 1 request per 5 seconds
});

export default RateLimiter;
