package com.whatisthis.scorer.consumers;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.whatisthis.scorer.constants.RabbitMQConstants;
import com.whatisthis.scorer.services.ScoreService;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import com.whatisthis.scorer.entities.Score;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.math.BigDecimal;

@Component
public class ScoreConsumer {

    @Autowired
    private ScoreService scoreService;

    @Autowired
    private RabbitTemplate rabbitTemplate;

    private final ObjectMapper objectMapper = new ObjectMapper();

    @RabbitListener(queues = RabbitMQConstants.CALCULATE_SCORE_QUEUE)
    public void consume(String message) {
        try {
            UserCreatedEvent userCreatedEvent = new UserCreatedEvent(message);
            System.out.println("Received message: " + userCreatedEvent);

            Score score = scoreService.calculateScore(
                    userCreatedEvent.id(),
                    BigDecimal.valueOf(userCreatedEvent.annual_income()),
                    BigDecimal.valueOf(userCreatedEvent.debt()),
                    BigDecimal.valueOf(userCreatedEvent.assets_value())
            );

            broadcastScoreUpdatedEvent(score);
        } catch (JsonProcessingException e) {
            System.out.println("Error parsing message: " + e.getMessage());
        }
    }

    private void broadcastScoreUpdatedEvent(Score score) {
        try {
            ScoreUpdatedEvent scoreUpdatedEvent = new ScoreUpdatedEvent(
                    score.getUserId(),
                    score.getCreditScore()
            );
            String message = objectMapper.writeValueAsString(scoreUpdatedEvent);
            rabbitTemplate.convertAndSend(RabbitMQConstants.SCORE_UPDATED_QUEUE, message);
            System.out.println("Broadcast score updated event: " + message);
        } catch (JsonProcessingException e) {
            System.out.println("Error broadcasting score updated event: " + e.getMessage());
        }
    }
}
