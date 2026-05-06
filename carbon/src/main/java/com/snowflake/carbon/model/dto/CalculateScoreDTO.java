package com.snowflake.carbon.model.dto;

import java.math.BigDecimal;

public record CalculateScoreDTO(
        BigDecimal income,
        BigDecimal debt,
        BigDecimal assetsValue
) {
}
