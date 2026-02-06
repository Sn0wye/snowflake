package com.snowflake.scorer.services;

import com.snowflake.scorer.exceptions.ScoreCalculationException;
import com.snowflake.scorer.model.dto.AssetScore;
import com.snowflake.scorer.model.dto.DebtScore;
import com.snowflake.scorer.model.dto.IncomeScore;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.math.RoundingMode;

@Service
public class CreditScoreService {
    public int calculateCreditScore(
            BigDecimal income,
            BigDecimal debt,
            BigDecimal assetsValue
    ) {
        IncomeScore incomeScore = this.calculateIncomeScore(income);
        DebtScore debtScore = this.calculateDebtScore(income, debt);
        AssetScore assetScore = this.calculateAssetScore(assetsValue);

        return incomeScore.getScore() + debtScore.getScore() + assetScore.getScore();
    }

    private static final BigDecimal EXPENSIVE_ASSETS_THRESHOLD = new BigDecimal("500000"); // $500,000.00
    private static final BigDecimal MEDIUM_ASSETS_THRESHOLD = new BigDecimal("200000");   // $200,000.00
    private static final int HIGH_ASSETS_SCORE = 300;
    private static final int MEDIUM_ASSETS_SCORE_MIN = 150;
    private static final int MEDIUM_ASSETS_SCORE_MAX = 300;
    private static final int LOW_ASSETS_SCORE = 0;

    public AssetScore calculateAssetScore(BigDecimal assetsValue) {
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
            BigDecimal proportion = assetsValue.subtract(MEDIUM_ASSETS_THRESHOLD).divide(EXPENSIVE_ASSETS_THRESHOLD.subtract(MEDIUM_ASSETS_THRESHOLD), 4, RoundingMode.HALF_EVEN);
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


    private static final BigDecimal LOW_DEBT_RATIO_THRESHOLD_PERCENT = new BigDecimal("20"); // 20%
    private static final BigDecimal MEDIUM_DEBT_RATIO_THRESHOLD_PERCENT = new BigDecimal("40"); // 40%
    private static final int HIGH_DEBT_RATIO_SCORE = 300;
    private static final int MEDIUM_DEBT_RATIO_SCORE_MIN = 150;
    private static final int MEDIUM_DEBT_RATIO_SCORE_MAX = 300;
    private static final int LOW_DEBT_RATIO_SCORE = 0;

    public DebtScore calculateDebtScore(BigDecimal income, BigDecimal debt) {
        if (income.compareTo(BigDecimal.ZERO) == 0) {
            throw new ScoreCalculationException("Income cannot be zero.");
        }

        // Calculate the ratio as a percentage (debt / income * 100)
        BigDecimal ratioPercent = debt.divide(income, 4, RoundingMode.HALF_EVEN).multiply(new BigDecimal("100"));

        // Debt exceeds income
        if (debt.compareTo(income) > 0) {
            return new DebtScore(LOW_DEBT_RATIO_SCORE); // Assign the minimum score or customize logic
        }

        if (ratioPercent.compareTo(LOW_DEBT_RATIO_THRESHOLD_PERCENT) < 0) {
            return new DebtScore(HIGH_DEBT_RATIO_SCORE);
        }

        if (ratioPercent.compareTo(MEDIUM_DEBT_RATIO_THRESHOLD_PERCENT) < 0) {
            BigDecimal proportion = ratioPercent.subtract(LOW_DEBT_RATIO_THRESHOLD_PERCENT).divide(MEDIUM_DEBT_RATIO_THRESHOLD_PERCENT.subtract(LOW_DEBT_RATIO_THRESHOLD_PERCENT), 4, RoundingMode.HALF_EVEN);
            int score = MEDIUM_DEBT_RATIO_SCORE_MIN + proportion.multiply(new BigDecimal(MEDIUM_DEBT_RATIO_SCORE_MAX - MEDIUM_DEBT_RATIO_SCORE_MIN)).intValue();
            return new DebtScore(score);
        }

        BigDecimal proportion = ratioPercent.subtract(MEDIUM_DEBT_RATIO_THRESHOLD_PERCENT).divide(new BigDecimal("100").subtract(MEDIUM_DEBT_RATIO_THRESHOLD_PERCENT), 4, RoundingMode.HALF_EVEN);
        int score = LOW_DEBT_RATIO_SCORE + proportion.multiply(new BigDecimal(MEDIUM_DEBT_RATIO_SCORE_MIN - LOW_DEBT_RATIO_SCORE)).intValue();
        return new DebtScore(score);
    }


    private static final BigDecimal HIGH_INCOME_THRESHOLD = new BigDecimal("100000"); // $100,000.00
    private static final BigDecimal MEDIUM_INCOME_THRESHOLD = new BigDecimal("50000"); // $50,000.00
    private static final int HIGH_INCOME_SCORE = 300;
    private static final int MEDIUM_INCOME_SCORE_MIN = 150;
    private static final int MEDIUM_INCOME_SCORE_MAX = 300;
    private static final int LOW_INCOME_SCORE = 0;

    public IncomeScore calculateIncomeScore(BigDecimal income) {
        // Check if income is greater than HIGH_INCOME_THRESHOLD
        if (income.compareTo(HIGH_INCOME_THRESHOLD) > 0) {
            return new IncomeScore(HIGH_INCOME_SCORE);
        }

        // Check if income is greater than MEDIUM_INCOME_THRESHOLD
        if (income.compareTo(MEDIUM_INCOME_THRESHOLD) > 0) {
            BigDecimal proportion = income.subtract(MEDIUM_INCOME_THRESHOLD).divide(HIGH_INCOME_THRESHOLD.subtract(MEDIUM_INCOME_THRESHOLD), 4, RoundingMode.HALF_EVEN);
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
