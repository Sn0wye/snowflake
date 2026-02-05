package com.whatisthis.scorer.consumers;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.util.Date;

public record UserCreatedEvent(
        String id,
        String username,
        String email,
        double annual_income,
        double debt,
        double assets_value,
        Date created_at
) {
    public UserCreatedEvent(String json) throws JsonProcessingException {
        this(new ObjectMapper().readValue(json, UserCreatedEvent.class));
    }

    private UserCreatedEvent(UserCreatedEvent userCreatedEvent) {
        this(userCreatedEvent.id(), userCreatedEvent.username(), userCreatedEvent.email(),
                userCreatedEvent.annual_income(), userCreatedEvent.debt(),
                userCreatedEvent.assets_value(), userCreatedEvent.created_at());
    }
}