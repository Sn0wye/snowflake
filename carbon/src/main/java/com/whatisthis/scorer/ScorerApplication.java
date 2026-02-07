package com.snowflake.scorer;

import io.swagger.v3.oas.annotations.OpenAPIDefinition;
import io.swagger.v3.oas.annotations.info.Contact;
import io.swagger.v3.oas.annotations.info.Info;
import io.swagger.v3.oas.annotations.info.License;
import io.swagger.v3.oas.annotations.servers.Server;
import org.springframework.amqp.rabbit.annotation.EnableRabbit;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
@EnableRabbit
@OpenAPIDefinition(
        info = @Info(
                title = "Snowflake API Reference",
                version = "1.0.0",
                description = "The Snowflake API is organized around REST. This API has predictable resource-oriented URLs, accepts JSON-encoded request bodies, returns JSON-encoded responses, and uses standard HTTP response codes, authentication, and verbs.",
                contact = @Contact(
                        name = "GitHub",
                        url = "https://github.com/Sn0wye/snowflake/issues"
                ),
                license = @License(
                        name = "GNU General Public License v3.0",
                        url = "https://www.gnu.org/licenses/gpl-3.0"
                )
        ),
        servers = {
                @Server(url = "https://snowflake.snowye.dev", description = "Snowflake API Server")
        }
)
public class ScorerApplication {
    public static void main(String[] args) {
        SpringApplication.run(ScorerApplication.class, args);
    }
}
