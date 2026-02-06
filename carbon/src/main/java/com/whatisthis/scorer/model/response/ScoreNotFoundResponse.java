package com.snowflake.scorer.model.response;

import io.swagger.v3.oas.annotations.media.Schema;

public record ScoreNotFoundResponse(
        @Schema(description = "Error message", example = "Score not found")
        String message,
        @Schema(description = "Status code", example = "404")
        Integer status_code
) {
}
