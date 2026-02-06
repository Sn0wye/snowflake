package com.snowflake.scorer.services;

import com.snowflake.scorer.exceptions.ScoreCalculationException;
import com.snowflake.scorer.model.dto.DebtScore;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.math.RoundingMode;

@Service
public class CalculateDebtScoreService {

    private static final BigDecimal LOW_RATIO_THRESHOLD_PERCENT = new BigDecimal("20"); // 20%
    private static final BigDecimal MEDIUM_RATIO_THRESHOLD_PERCENT = new BigDecimal("40"); // 40%
    private static final int HIGH_RATIO_SCORE = 300;
    private static final int MEDIUM_RATIO_SCORE_MIN = 150;
    private static final int MEDIUM_RATIO_SCORE_MAX = 300;
    private static final int LOW_RATIO_SCORE = 0;

    public DebtScore calculate(BigDecimal income, BigDecimal debt) {
        if (income.compareTo(BigDecimal.ZERO) == 0) {
            throw new ScoreCalculationException("Income cannot be zero.");
        }

        // Calculate the ratio as a percentage (debt / income * 100)
        BigDecimal ratioPercent = debt.divide(income, 4, RoundingMode.HALF_EVEN).multiply(new BigDecimal("100"));

        // Debt exceeds income
        if (debt.compareTo(income) > 0) {
            return new DebtScore(LOW_RATIO_SCORE); // Assign the minimum score or customize logic
        }

        if (ratioPercent.compareTo(LOW_RATIO_THRESHOLD_PERCENT) < 0) {
            return new DebtScore(HIGH_RATIO_SCORE);
        }

        if (ratioPercent.compareTo(MEDIUM_RATIO_THRESHOLD_PERCENT) < 0) {
            BigDecimal proportion = ratioPercent.subtract(LOW_RATIO_THRESHOLD_PERCENT)
                    .divide(MEDIUM_RATIO_THRESHOLD_PERCENT.subtract(LOW_RATIO_THRESHOLD_PERCENT), 4, RoundingMode.HALF_EVEN);
            int score = MEDIUM_RATIO_SCORE_MIN + proportion.multiply(new BigDecimal(MEDIUM_RATIO_SCORE_MAX - MEDIUM_RATIO_SCORE_MIN)).intValue();
            return new DebtScore(score);
        }

        BigDecimal proportion = ratioPercent.subtract(MEDIUM_RATIO_THRESHOLD_PERCENT)
                .divide(new BigDecimal("100").subtract(MEDIUM_RATIO_THRESHOLD_PERCENT), 4, RoundingMode.HALF_EVEN);
        int score = LOW_RATIO_SCORE + proportion.multiply(new BigDecimal(MEDIUM_RATIO_SCORE_MIN - LOW_RATIO_SCORE)).intValue();
        return new DebtScore(score);
    }
}
