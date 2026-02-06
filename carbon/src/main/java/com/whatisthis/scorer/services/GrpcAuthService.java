package com.snowflake.scorer.services;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.Status;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import pb.Auth;
import pb.AuthServiceGrpc;

@Service
public class GrpcAuthService {

    private final AuthServiceGrpc.AuthServiceBlockingStub authServiceStub;

    public GrpcAuthService(
            @Value("${spring.grpc.host}") String host,
            @Value("${spring.grpc.port}") int port
    ) {
        ManagedChannel channel = ManagedChannelBuilder.forAddress(host, port)
                .usePlaintext()
                .build();

        authServiceStub = AuthServiceGrpc.newBlockingStub(channel);
    }

    public boolean validateToken(String token) {
        Auth.ValidateTokenRequest request = Auth.ValidateTokenRequest.newBuilder().setToken(token).build();
        try {
            Auth.ValidateTokenResponse response = authServiceStub.validateToken(request);
            return response.getValid();
        } catch (Exception e) {
            if (Status.fromThrowable(e).getCode() == Status.Code.UNAVAILABLE) {
                System.out.println("gRPC server is unavailable!");
                return false;
            }
            throw new RuntimeException("Failed to validate token", e);
        }
    }

    public Auth.ParseTokenResponse parseToken(String token) throws Exception {
        Auth.ParseTokenRequest request = Auth.ParseTokenRequest.newBuilder().setToken(token).build();
        try {
            return authServiceStub.parseToken(request);
        } catch (Exception e) {
            if (Status.fromThrowable(e).getCode() == Status.Code.UNAVAILABLE) {
                System.out.println("gRPC server is unavailable!");
                return null;
            }
            throw new Exception("Failed to parse token", e);
        }
    }
}
