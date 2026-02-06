package com.snowflake.scorer.model.response;

import io.swagger.v3.oas.annotations.media.Schema;

import java.util.Map;

public record ValidationErrorResponse(
        @Schema(description = "Error message", example = "Invalid request body.")
        String message,
        @Schema(description = "Status code", example = "400")
        Integer status_code,
        @Schema(description = "Validation errors", example = "{\"income\":\"Income cannot be null.\",\"debt\":\"Debt cannot be negative.\"}")
        Map<String, String> errors
) {
}
