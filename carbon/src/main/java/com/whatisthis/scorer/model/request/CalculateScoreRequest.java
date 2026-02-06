package com.snowflake.scorer.model.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotNull;

import java.math.BigDecimal;
import java.math.RoundingMode;

public record CalculateScoreRequest(
        @Schema(description = "Income", example = "5000.00")
        @NotNull(message = "Income cannot be null.")
        @Min(value = 1, message = "Income must be greater than zero.")
        Double income,

        @Schema(description = "Debt", example = "1000.00")
        @NotNull(message = "Debt cannot be null.")
        @Min(value = 0, message = "Debt cannot be negative.")
        Double debt,

        @Schema(description = "Assets value", example = "50000.00")
        @NotNull(message = "Assets value cannot be null.")
        @Min(value = 0, message = "Assets value cannot be negative.")
        Double assetsValue
) {
    public BigDecimal incomeAsDecimal() {
        return BigDecimal.valueOf(income).setScale(4, RoundingMode.HALF_EVEN);
    }

    public BigDecimal debtAsDecimal() {
        return BigDecimal.valueOf(debt).setScale(4, RoundingMode.HALF_EVEN);
    }

    public BigDecimal assetsValueAsDecimal() {
        return BigDecimal.valueOf(assetsValue).setScale(4, RoundingMode.HALF_EVEN);
    }
}
