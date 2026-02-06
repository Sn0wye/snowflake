package com.snowflake.scorer.services;

import com.snowflake.scorer.model.dto.IncomeScore;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.math.RoundingMode;

@Service
public class CalculateIncomeScoreService {

    private static final BigDecimal HIGH_INCOME_THRESHOLD = new BigDecimal("100000"); // $100,000.00
    private static final BigDecimal MEDIUM_INCOME_THRESHOLD = new BigDecimal("50000"); // $50,000.00
    private static final int HIGH_INCOME_SCORE = 300;
    private static final int MEDIUM_INCOME_SCORE_MIN = 150;
    private static final int MEDIUM_INCOME_SCORE_MAX = 300;
    private static final int LOW_INCOME_SCORE = 0;

    public IncomeScore calculate(BigDecimal income) {
        // Check if income is greater than HIGH_INCOME_THRESHOLD
        if (income.compareTo(HIGH_INCOME_THRESHOLD) > 0) {
            return new IncomeScore(HIGH_INCOME_SCORE);
        }

        // Check if income is greater than MEDIUM_INCOME_THRESHOLD
        if (income.compareTo(MEDIUM_INCOME_THRESHOLD) > 0) {
            BigDecimal proportion = income.subtract(MEDIUM_INCOME_THRESHOLD)
                    .divide(HIGH_INCOME_THRESHOLD.subtract(MEDIUM_INCOME_THRESHOLD), 4, RoundingMode.HALF_EVEN);
            int score = MEDIUM_INCOME_SCORE_MIN + proportion.multiply(new BigDecimal(MEDIUM_INCOME_SCORE_MAX - MEDIUM_INCOME_SCORE_MIN)).intValue();
            return new IncomeScore(score);
        }

        // If income is above zero but less than MEDIUM_INCOME_THRESHOLD
        if (income.compareTo(BigDecimal.ZERO) > 0) {
            BigDecimal proportion = income.divide(MEDIUM_INCOME_THRESHOLD, 4, RoundingMode.HALF_EVEN);
            int score = proportion.multiply(new BigDecimal(MEDIUM_INCOME_SCORE_MIN)).intValue();
            return new IncomeScore(score);
        }

        return new IncomeScore(LOW_INCOME_SCORE);
    }
}
