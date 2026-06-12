package com.snowflake.carbon.consumers;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.snowflake.carbon.constants.RabbitMQConstants;
import com.snowflake.carbon.services.ScoreService;
import org.slf4j.MDC;
import org.springframework.amqp.core.Message;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import com.snowflake.carbon.entities.Score;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.math.BigDecimal;
import java.util.UUID;

@Component
public class ScoreConsumer {

    private static final Logger log = LoggerFactory.getLogger(ScoreConsumer.class);

    @Autowired
    private ScoreService scoreService;

    @Autowired
    private RabbitTemplate rabbitTemplate;

    private final ObjectMapper objectMapper = new ObjectMapper();

    @RabbitListener(queues = RabbitMQConstants.CALCULATE_SCORE_QUEUE)
    public void consume(Message message) {
        String correlationId = extractCorrelationId(message);
        MDC.put("correlation_id", correlationId);

        try {
            String body = new String(message.getBody());
            UserCreatedEvent userCreatedEvent = new UserCreatedEvent(body);
            log.info("Received message: {}", userCreatedEvent);

            Score score = scoreService.calculateScore(
                    userCreatedEvent.id(),
                    BigDecimal.valueOf(userCreatedEvent.annual_income()),
                    BigDecimal.valueOf(userCreatedEvent.debt()),
                    BigDecimal.valueOf(userCreatedEvent.assets_value())
            );

            broadcastScoreUpdatedEvent(score);
        } catch (JsonProcessingException e) {
            log.error("Error parsing message: {}", e.getMessage());
        } finally {
            MDC.remove("correlation_id");
        }
    }

    private void broadcastScoreUpdatedEvent(Score score) {
        try {
            ScoreUpdatedEvent scoreUpdatedEvent = new ScoreUpdatedEvent(
                    score.getUserId(),
                    score.getCreditScore()
            );
            String msg = objectMapper.writeValueAsString(scoreUpdatedEvent);
            rabbitTemplate.convertAndSend(RabbitMQConstants.SCORE_UPDATED_QUEUE, msg);
            log.info("Broadcast score updated event: {}", msg);
        } catch (JsonProcessingException e) {
            log.error("Error broadcasting score updated event: {}", e.getMessage());
        }
    }

    private String extractCorrelationId(Message message) {
        Object headers = message.getMessageProperties().getHeaders().get("x-correlation-id");
        if (headers instanceof String s && !s.isEmpty()) {
            return s;
        }
        return UUID.randomUUID().toString();
    }
}
