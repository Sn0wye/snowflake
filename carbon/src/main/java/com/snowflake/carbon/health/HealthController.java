package com.snowflake.carbon.health;

import java.sql.Connection;
import java.util.HashMap;
import java.util.Map;

import javax.sql.DataSource;

import org.springframework.amqp.rabbit.connection.ConnectionFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/score")
public class HealthController {

    @Autowired
    private DataSource dataSource;

    @Autowired
    private ConnectionFactory rabbitConnectionFactory;

    @GetMapping("/health")
    public ResponseEntity<Map<String, Object>> health() {
        Map<String, String> components = new HashMap<>();
        boolean allHealthy = true;

        // Database check
        try (Connection conn = dataSource.getConnection()) {
            if (conn.isValid(5)) {
                components.put("database", "ok");
            } else {
                components.put("database", "connection invalid");
                allHealthy = false;
            }
        } catch (Exception e) {
            components.put("database", e.getMessage());
            allHealthy = false;
        }

        // RabbitMQ check
        try {
            var conn = rabbitConnectionFactory.createConnection();
            var channel = conn.createChannel(false);
            channel.close();
            conn.close();
            components.put("rabbitmq", "ok");
        } catch (Exception e) {
            components.put("rabbitmq", e.getMessage());
            allHealthy = false;
        }

        Map<String, Object> response = new HashMap<>();
        response.put("status", allHealthy ? "healthy" : "unhealthy");
        response.put("components", components);

        HttpStatus httpStatus = allHealthy ? HttpStatus.OK : HttpStatus.SERVICE_UNAVAILABLE;
        return ResponseEntity.status(httpStatus).body(response);
    }
}
