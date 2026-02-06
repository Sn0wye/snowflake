package com.snowflake.scorer.model.response;

import io.swagger.v3.oas.annotations.media.Schema;

public record ScoreResponse(
        @Schema(description = "Credit score", example = "700")
        int creditScore
) {
}
