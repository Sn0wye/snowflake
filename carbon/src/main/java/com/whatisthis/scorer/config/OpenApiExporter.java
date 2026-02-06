package com.snowflake.scorer.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.ApplicationRunner;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.client.RestTemplate;

import java.io.FileWriter;
import java.io.IOException;

@Configuration
public class OpenApiExporter {

    @Value("${openapi.output.path:openapi.json}")
    private String outputPath;

    @Bean
    public ApplicationRunner saveOpenApiSpec() {
        return args -> {
            String apiDocsUrl = "http://localhost:8081/v3/api-docs"; // OpenAPI JSON endpoint

            try {
                RestTemplate restTemplate = new RestTemplate();
                String openApiJson = restTemplate.getForObject(apiDocsUrl, String.class);

                if (openApiJson == null) {
                    System.err.println("Failed to fetch OpenAPI spec");
                    return;
                }

                try (FileWriter writer = new FileWriter(outputPath)) {
                    writer.write(openApiJson);
                }

                System.out.println("OpenAPI spec saved to: " + outputPath);
            } catch (IOException e) {
                System.err.println("Failed to save OpenAPI spec: " + e.getMessage());
            }
        };
    }
}
