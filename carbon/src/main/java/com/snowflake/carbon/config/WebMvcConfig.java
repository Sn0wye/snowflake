package com.snowflake.carbon.config;

import com.snowflake.carbon.interceptors.BearerTokenInterceptor;
import com.snowflake.carbon.interceptors.CorrelationIdInterceptor;
import com.snowflake.carbon.interceptors.RateLimitInterceptor;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration
public class WebMvcConfig implements WebMvcConfigurer {

    @Autowired
    private CorrelationIdInterceptor correlationIdInterceptor;

    @Autowired
    private BearerTokenInterceptor bearerTokenInterceptor;

    @Autowired
    private RateLimitInterceptor rateLimitInterceptor;

    @Override
    public void addInterceptors(InterceptorRegistry registry) {
        // Actuator endpoints (/metrics scraped by Prometheus, /health) are reached
        // without a bearer token, so they bypass auth/rate-limit interceptors.
        registry.addInterceptor(correlationIdInterceptor)
                .excludePathPatterns("/score/health", "/metrics", "/health");
        registry.addInterceptor(bearerTokenInterceptor)
                .excludePathPatterns("/score/health", "/metrics", "/health");
        registry.addInterceptor(rateLimitInterceptor)
                .excludePathPatterns("/score/health", "/metrics", "/health");
    }
}