package com.snowflake.scorer.model.response;

import io.swagger.v3.oas.annotations.media.Schema;

public record ErrorResponse(
        @Schema(description = "Error message", example = "Something went wrong :(")
        String message,
        @Schema(description = "Status code", example = "500")
        Integer status_code
) {
}
