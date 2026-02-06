package com.snowflake.scorer.exceptions;

public class ScoreNotFoundException extends RuntimeException {
    public ScoreNotFoundException(String message) {
        super(message);
    }
}
