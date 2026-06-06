package com.snowflake.carbon.interceptors;

import com.snowflake.carbon.exceptions.TooManyRequestsException;
import jakarta.annotation.PreDestroy;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;

import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;

@Component
public class RateLimitInterceptor implements HandlerInterceptor {

    private final ConcurrentHashMap<String, FixedWindow> buckets = new ConcurrentHashMap<>();
    private final ScheduledExecutorService reapScheduler = Executors.newSingleThreadScheduledExecutor(r -> {
        Thread t = new Thread(r, "ratelimit-reaper");
        t.setDaemon(true);
        return t;
    });

    @Value("${ratelimit.enabled:true}")
    private boolean enabled;

    @Value("${ratelimit.score-per-second:10}")
    private int rate;

    public RateLimitInterceptor() {
        reapScheduler.scheduleWithFixedDelay(this::reap, 1, 1, TimeUnit.MINUTES);
    }

    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) {
        if (!enabled) {
            return true;
        }

        String userId = (String) request.getAttribute("userId");
        if (userId == null) {
            return true;
        }

        FixedWindow window = buckets.computeIfAbsent(userId, k -> new FixedWindow());
        window.touch();
        if (!window.allow(rate)) {
            throw new TooManyRequestsException("Too many requests. Try again in 1 second.");
        }

        return true;
    }

    private void reap() {
        long cutoff = System.currentTimeMillis() - TimeUnit.MINUTES.toMillis(1);
        buckets.entrySet().removeIf(entry -> entry.getValue().lastSeen.get() < cutoff);
    }

    @PreDestroy
    public void shutdown() {
        reapScheduler.shutdownNow();
    }

    private static class FixedWindow {
        final AtomicLong lastSeen = new AtomicLong();
        private long windowStart = System.nanoTime();
        private int count;

        FixedWindow() {
            touch();
        }

        void touch() {
            lastSeen.set(System.currentTimeMillis());
        }

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
