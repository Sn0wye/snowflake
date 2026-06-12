package com.snowflake.carbon.config;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.ApplicationRunner;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.client.RestTemplate;

import java.io.FileWriter;
import java.io.IOException;

@Configuration
public class OpenApiExporter {

    private static final Logger log = LoggerFactory.getLogger(OpenApiExporter.class);

    @Value("${openapi.output.path:openapi.json}")
    private String outputPath;

    @Bean
    @ConditionalOnProperty(name = "openapi.export.enabled", havingValue = "true", matchIfMissing = true)
    public ApplicationRunner saveOpenApiSpec() {
        return args -> {
            String apiDocsUrl = "http://localhost:8081/v3/api-docs"; // OpenAPI JSON endpoint

            try {
                RestTemplate restTemplate = new RestTemplate();
                String openApiJson = restTemplate.getForObject(apiDocsUrl, String.class);

                if (openApiJson == null) {
                    log.warn("Failed to fetch OpenAPI spec");
                    return;
                }

                try (FileWriter writer = new FileWriter(outputPath)) {
                    writer.write(openApiJson);
                }

                log.info("OpenAPI spec saved to: {}", outputPath);
            } catch (IOException e) {
                log.error("Failed to save OpenAPI spec: {}", e.getMessage());
            }
        };
    }
}
