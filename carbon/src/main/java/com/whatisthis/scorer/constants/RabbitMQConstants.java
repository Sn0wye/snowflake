package com.snowflake.scorer.constants;

public final class RabbitMQConstants {

    // Private constructor to prevent instantiation
    private RabbitMQConstants() {
        throw new UnsupportedOperationException("This is a constants class and cannot be instantiated");
    }

    // Routing keys
    public static final String CALCULATE_SCORE_QUEUE = "user.created";
    public static final String SCORE_UPDATED_QUEUE = "score-updated";
}