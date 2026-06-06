package com.snowflake.carbon.interceptors;

import com.snowflake.carbon.exceptions.TooManyRequestsException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Component
public class RateLimitInterceptor implements HandlerInterceptor {

    private final Map<String, SlidingWindow> buckets = new ConcurrentHashMap<>();

    @Value("${ratelimit.enabled:true}")
    private boolean enabled;

    @Value("${ratelimit.score-per-second:10}")
    private int rate;

    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) {
        if (!enabled) {
            return true;
        }

        String userId = (String) request.getAttribute("userId");
        if (userId == null) {
            return true;
        }

        SlidingWindow window = buckets.computeIfAbsent(userId, k -> new SlidingWindow());
        if (!window.allow(rate)) {
            throw new TooManyRequestsException("Too many requests. Try again in 1 second.");
        }

        return true;
    }

    private static class SlidingWindow {
        private long windowStart = System.nanoTime();
        private int count;

        synchronized boolean allow(int rate) {
            long now = System.nanoTime();
            if (now - windowStart > 1_000_000_000L) {
                windowStart = now;
                count = 0;
            }
            if (count >= rate) {
                return false;
            }
            count++;
            return true;
        }
    }
}
