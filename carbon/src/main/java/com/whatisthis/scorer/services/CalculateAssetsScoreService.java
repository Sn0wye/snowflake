package com.snowflake.scorer.services;

import com.snowflake.scorer.exceptions.ScoreCalculationException;
import com.snowflake.scorer.model.dto.AssetScore;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.math.RoundingMode;

@Service
public class CalculateAssetsScoreService {

    private static final BigDecimal EXPENSIVE_ASSETS_THRESHOLD = new BigDecimal("500000"); // $500,000.00
    private static final BigDecimal MEDIUM_ASSETS_THRESHOLD = new BigDecimal("200000");   // $200,000.00
    private static final int HIGH_ASSETS_SCORE = 300;
    private static final int MEDIUM_ASSETS_SCORE_MIN = 150;
    private static final int MEDIUM_ASSETS_SCORE_MAX = 300;
    private static final int LOW_ASSETS_SCORE = 0;

    public AssetScore calculate(BigDecimal assetsValue) {
        // Ensure that the input value is not negative
        if (assetsValue.compareTo(BigDecimal.ZERO) < 0) {
            throw new ScoreCalculationException("Assets value cannot be negative.");
        }

        // Check if assets value is greater than the expensive assets threshold
        if (assetsValue.compareTo(EXPENSIVE_ASSETS_THRESHOLD) > 0) {
            return new AssetScore(HIGH_ASSETS_SCORE);
        }

        // Check if assets value is greater than the medium assets threshold
        if (assetsValue.compareTo(MEDIUM_ASSETS_THRESHOLD) > 0) {
            BigDecimal proportion = assetsValue.subtract(MEDIUM_ASSETS_THRESHOLD)
                    .divide(EXPENSIVE_ASSETS_THRESHOLD.subtract(MEDIUM_ASSETS_THRESHOLD), 4, RoundingMode.HALF_EVEN);
            int score = MEDIUM_ASSETS_SCORE_MIN + proportion.multiply(new BigDecimal(MEDIUM_ASSETS_SCORE_MAX - MEDIUM_ASSETS_SCORE_MIN)).intValue();
            return new AssetScore(score);
        }

        // Check if assets value is positive but less than medium assets threshold
        if (assetsValue.compareTo(BigDecimal.ZERO) > 0) {
            BigDecimal proportion = assetsValue.divide(MEDIUM_ASSETS_THRESHOLD, 4, RoundingMode.HALF_EVEN);
            int score = proportion.multiply(new BigDecimal(MEDIUM_ASSETS_SCORE_MIN)).intValue();
            return new AssetScore(score);
        }

        return new AssetScore(LOW_ASSETS_SCORE);
    }
}
