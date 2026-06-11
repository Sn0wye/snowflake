package com.snowflake.carbon.services;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import jakarta.annotation.PreDestroy;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import pb.Auth;
import pb.AuthServiceGrpc;

import java.util.Map;
import java.util.concurrent.TimeUnit;

@Service
public class GrpcAuthService {

    private final ManagedChannel channel;
    private final AuthServiceGrpc.AuthServiceBlockingStub authServiceStub;

    public GrpcAuthService(
            @Value("${spring.grpc.host}") String host,
            @Value("${spring.grpc.port}") int port
    ) {
        this.channel = ManagedChannelBuilder.forAddress(host, port)
                .usePlaintext()
                .defaultServiceConfig(Map.of(
                        "methodConfig", java.util.List.of(Map.of(
                                "name", java.util.List.of(Map.of("service", "pb.AuthService")),
                                "retryPolicy", Map.of(
                                        // gRPC service config maps only accept Double, not Integer
                                        "maxAttempts", 4.0,
                                        "initialBackoff", "0.5s",
                                        "maxBackoff", "5s",
                                        "backoffMultiplier", 2.0,
                                        "retryableStatusCodes", java.util.List.of("UNAVAILABLE")
                                )
                        ))
                ))
                .enableRetry()
                .build();

        this.authServiceStub = AuthServiceGrpc.newBlockingStub(channel);
    }

    @PreDestroy
    private void shutdown() {
        channel.shutdown();
        try {
            if (!channel.awaitTermination(5, TimeUnit.SECONDS)) {
                channel.shutdownNow();
            }
        } catch (InterruptedException e) {
            channel.shutdownNow();
            Thread.currentThread().interrupt();
        }
    }

    public boolean validateToken(String token) {
        Auth.ValidateTokenRequest request = Auth.ValidateTokenRequest.newBuilder().setToken(token).build();
        Auth.ValidateTokenResponse response = authServiceStub.validateToken(request);
        return response.getValid();
    }

    public Auth.ParseTokenResponse parseToken(String token) {
        Auth.ParseTokenRequest request = Auth.ParseTokenRequest.newBuilder().setToken(token).build();
        return authServiceStub.parseToken(request);
    }
}
