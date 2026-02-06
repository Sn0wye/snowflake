package com.snowflake.scorer.interceptors;

import com.snowflake.scorer.exceptions.UnauthorizedException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;
import com.snowflake.scorer.services.GrpcAuthService;
import pb.Auth;

// This is an overengineered way to validate JWT tokens.
// The token is validated by a gRPC service.
// Ideally, the token should be validated by the application itself.
// This will be this way until I find some use for the Authentication service gRPC :D
@Component
public class BearerTokenInterceptor implements HandlerInterceptor {

    @Autowired
    private GrpcAuthService grpcAuthService;

    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) {
        String requestUri = request.getRequestURI();

        // Ignore Swagger routes from token validation
        if (requestUri.startsWith("/swagger") ||
                requestUri.startsWith("/v3/api-docs") ||
                requestUri.equals("/swagger.html")) {
            return true;
        }

        String authorizationHeader = request.getHeader("Authorization");
        if (authorizationHeader != null && authorizationHeader.startsWith("Bearer ")) {
            String bearerToken = authorizationHeader.substring(7); // Remove "Bearer " prefix
            try {
                Auth.ParseTokenResponse validatedToken = grpcAuthService.parseToken(bearerToken);

                if (validatedToken == null) {
                    System.err.println("Token validation returned null");
                    throw new UnauthorizedException("Invalid token");
                }

                request.setAttribute("bearerToken", bearerToken);
                request.setAttribute("userId", validatedToken.getSub());
                request.setAttribute("token", validatedToken);
            } catch (Exception e) {
                throw new UnauthorizedException("JWT parsing or validation failed" + e.getMessage());
            }
        } else {
            throw new UnauthorizedException("Unauthorized");
        }
        return true;
    }

}