package com.snowflake.scorer.services;

import com.snowflake.scorer.exceptions.ScoreCalculationException;
import com.snowflake.scorer.model.dto.IncomeScore;
import com.snowflake.scorer.model.dto.DebtScore;
import com.snowflake.scorer.model.dto.AssetScore;
import org.junit.jupiter.api.Test;

import java.math.BigDecimal;
import java.math.RoundingMode;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.junit.jupiter.api.Assertions.assertThrows;

class CreditScoreServiceTest {

    private final CreditScoreService creditScoreService = new CreditScoreService();

    @Test
    void shouldReturnSumOfAllScores() {
        // Given
        BigDecimal income = BigDecimal.valueOf(100_000);
        BigDecimal debt = BigDecimal.valueOf(20_000);
        BigDecimal assetsValue = BigDecimal.valueOf(200_000);

        // When
        int result = creditScoreService.calculateCreditScore(income, debt, assetsValue);

        // Then
        assertThat(result).isEqualTo(300 + 150 + 150); // High income, medium debt, medium asset score
    }

    @Test
    void shouldHandleZeroValuesCorrectly() {
        // Given
        BigDecimal income = BigDecimal.ZERO;
        BigDecimal debt = BigDecimal.ZERO;
        BigDecimal assetsValue = BigDecimal.ZERO;

        // When
        assertThrows(ScoreCalculationException.class, () -> creditScoreService.calculateCreditScore(income, debt, assetsValue));
    }

    @Test
    void shouldCalculateCorrectlyWhenIncomeScoreDominates() {
        // Given
        BigDecimal income = BigDecimal.valueOf(200_000);
        BigDecimal debt = BigDecimal.valueOf(10_000);
        BigDecimal assetsValue = BigDecimal.valueOf(50_000);

        // When
        int result = creditScoreService.calculateCreditScore(income, debt, assetsValue);

        // Then
        assertThat(result).isEqualTo(300 + 300 + 37); // High income, high debt score, low asset score
    }

    @Test
    void shouldCalculateAssetScoreForHighValue() {
        // Given
        BigDecimal assetsValue = BigDecimal.valueOf(600_000);

        // When
        AssetScore assetScore = creditScoreService.calculateAssetScore(assetsValue);

        // Then
        assertThat(assetScore.getScore()).isEqualTo(300); // Maximum asset score
    }

    @Test
    void shouldCalculateDebtScoreForLowDebtRatio() {
        // Given
        BigDecimal income = BigDecimal.valueOf(100_000);
        BigDecimal debt = BigDecimal.valueOf(10_000);

        // When
        DebtScore debtScore = creditScoreService.calculateDebtScore(income, debt);

        // Then
        assertThat(debtScore.getScore()).isEqualTo(300); // Debt ratio well below threshold
    }

    @Test
    void shouldThrowExceptionForZeroIncomeInDebtScoreCalculation() {
        // Given
        BigDecimal income = BigDecimal.ZERO;
        BigDecimal debt = BigDecimal.valueOf(10_000);

        // When & Then
        assertThrows(ScoreCalculationException.class, () -> creditScoreService.calculateDebtScore(income, debt));
    }

    @Test
    void shouldCalculateIncomeScoreForMediumIncome() {
        // Given
        BigDecimal income = BigDecimal.valueOf(75_000);

        // When
        IncomeScore incomeScore = creditScoreService.calculateIncomeScore(income);

        // Then
        assertThat(incomeScore.getScore()).isEqualTo(225); // Proportional medium income score
    }

    @Test
    void shouldCalculateCreditScoreWithExactValues() {
        // Given
        BigDecimal income = BigDecimal.valueOf(80_000);
        BigDecimal debt = BigDecimal.valueOf(16_000);
        BigDecimal assetsValue = BigDecimal.valueOf(301_000);

        // When
        int creditScore = creditScoreService.calculateCreditScore(income, debt, assetsValue);

        // Then
        assertThat(creditScore).isEqualTo(240 + 150 + 200); // Income: 240, Debt: 150, Assets: 225
    }

    // Asset score calculation tests

    @Test
    void shouldReturnHighAssetsScoreForExpensiveAssets() {
        BigDecimal assetsValue = new BigDecimal("600000"); // Above the expensive assets threshold
        AssetScore result = creditScoreService.calculateAssetScore(assetsValue);
        assertThat(result.getScore()).isEqualTo(300);
    }

    @Test
    void shouldReturnMediumAssetsScoreWithinMediumRange() {
        BigDecimal assetsValue = new BigDecimal("350000"); // Between 200K and 500K
        AssetScore result = creditScoreService.calculateAssetScore(assetsValue);

        // Proportionally calculated score
        BigDecimal proportion = assetsValue.subtract(new BigDecimal("200000"))
                .divide(new BigDecimal("500000").subtract(new BigDecimal("200000")), 4, RoundingMode.HALF_EVEN);
        int expectedScore = 150 + proportion.multiply(new BigDecimal(300 - 150)).intValue();

        assertThat(result.getScore()).isEqualTo(expectedScore);
    }

    @Test
    void shouldReturnLowAssetsScoreForNoAssets() {
        BigDecimal assetsValue = new BigDecimal("0"); // No assets
        AssetScore result = creditScoreService.calculateAssetScore(assetsValue);
        assertThat(result.getScore()).isEqualTo(0);
    }

    @Test
    void shouldReturnZeroForNegativeAssetsValue() {
        BigDecimal assetsValue = new BigDecimal("-1"); // Negative assets

        assertThatThrownBy(() -> creditScoreService.calculateAssetScore(assetsValue))
                .isInstanceOf(ScoreCalculationException.class)
                .hasMessage("Assets value cannot be negative.");
    }

    @Test
    void shouldReturnScoreProportionalToLowAssetsValue() {
        BigDecimal assetsValue = new BigDecimal("100000"); // Below 200K threshold
        AssetScore result = creditScoreService.calculateAssetScore(assetsValue);

        // Proportionally calculated score
        BigDecimal proportion = assetsValue.divide(new BigDecimal("200000"), 4, RoundingMode.HALF_EVEN);
        int expectedScore = proportion.multiply(new BigDecimal(150)).intValue();

        assertThat(result.getScore()).isEqualTo(expectedScore);
    }

    @Test
    void shouldReturnMediumScoreForBoundaryOfMediumAssets() {
        BigDecimal assetsValue = new BigDecimal("200000"); // Exactly on the medium threshold
        AssetScore result = creditScoreService.calculateAssetScore(assetsValue);

        assertThat(result.getScore()).isEqualTo(150); // Lower bound of medium range
    }

    @Test
    void shouldReturnHighScoreForBoundaryOfExpensiveAssets() {
        BigDecimal assetsValue = new BigDecimal("500000"); // Exactly on the expensive threshold
        AssetScore result = creditScoreService.calculateAssetScore(assetsValue);

        assertThat(result.getScore()).isEqualTo(300); // Upper bound of medium range
    }

    // Debt score calculation tests

    @Test
    void shouldReturnHighRatioScoreForLowDebtRatio() {
        // Given
        BigDecimal income = BigDecimal.valueOf(100_000.00); // 100K
        BigDecimal debt = BigDecimal.valueOf(10_000.00);    // 10K (10% ratio)

        // When
        DebtScore result = creditScoreService.calculateDebtScore(income, debt);

        // Then
        assertThat(result.getScore()).isEqualTo(300); // High ratio score
    }

    @Test
    void shouldReturnProportionalScoreForMediumDebtRatio() {
        // Given
        BigDecimal income = BigDecimal.valueOf(100_000.00); // 100K
        BigDecimal debt = BigDecimal.valueOf(30_000.00);    // 30K (30% ratio)

        // When
        DebtScore result = creditScoreService.calculateDebtScore(income, debt);

        // Proportional score calculation
        double ratioPercent = 30.0; // 30% debt-to-income ratio
        double proportion = (ratioPercent - 20) / (40 - 20);
        int expectedScore = (int) (150 + proportion * (300 - 150));

        // Then
        assertThat(result.getScore()).isEqualTo(expectedScore);
    }

    @Test
    void shouldReturnLowRatioScoreForHighDebtRatio() {
        // Given
        BigDecimal income = BigDecimal.valueOf(100_000.00); // 100K
        BigDecimal debt = BigDecimal.valueOf(80_000.00);    // 80K (80% ratio)

        // When
        DebtScore result = creditScoreService.calculateDebtScore(income, debt);

        // Proportional score calculation
        double ratioPercent = 80.0; // 80% debt-to-income ratio
        double proportion = (ratioPercent - 40) / (100 - 40);
        int expectedScore = (int) (0 + proportion * 150);

        // Then
        assertThat(result.getScore()).isEqualTo(expectedScore);
    }

    @Test
    void shouldHandleDebtRatioAtLowThresholdBoundary() {
        // Given
        BigDecimal income = BigDecimal.valueOf(100_000.00); // 100K
        BigDecimal debt = BigDecimal.valueOf(20_000.00);    // 20K (20% ratio)

        // When
        DebtScore result = creditScoreService.calculateDebtScore(income, debt);

        // Then
        assertThat(result.getScore()).isEqualTo(150); // Lower bound of medium range
    }

    @Test
    void shouldHandleDebtRatioAtMediumThresholdBoundary() {
        // Given
        BigDecimal income = BigDecimal.valueOf(100_000.00); // 100K
        BigDecimal debt = BigDecimal.valueOf(40_000.00);    // 40K (40% ratio)

        // When
        DebtScore result = creditScoreService.calculateDebtScore(income, debt);

        // Then
        assertThat(result.getScore()).isEqualTo(0); // Upper bound of low range
    }

    @Test
    void shouldThrowExceptionWhenIncomeIsZero() {
        // Given
        BigDecimal income = BigDecimal.ZERO;
        BigDecimal debt = BigDecimal.valueOf(10_000.00); // 10K

        // Then
        assertThatThrownBy(() -> creditScoreService.calculateDebtScore(income, debt))
                .isInstanceOf(ScoreCalculationException.class)
                .hasMessage("Income cannot be zero.");
    }

    @Test
    void shouldReturnHighRatioScoreForZeroDebt() {
        // Given
        BigDecimal income = BigDecimal.valueOf(100_000.00); // 100K
        BigDecimal debt = BigDecimal.ZERO;                  // No debt (0% ratio)

        // When
        DebtScore result = creditScoreService.calculateDebtScore(income, debt);

        // Then
        assertThat(result.getScore()).isEqualTo(300); // High ratio score
    }

    // Income score calculation tests

    @Test
    void shouldReturnHighIncomeScoreForIncomeAboveHighThreshold() {
        // Given
        BigDecimal income = new BigDecimal("150000"); // 150K (income threshold)

        // When
        IncomeScore result = creditScoreService.calculateIncomeScore(income);

        // Then
        assertThat(result.getScore()).isEqualTo(300); // High income score
    }

    @Test
    void shouldReturnProportionalScoreForMediumIncomeRange() {
        // Given
        BigDecimal income = new BigDecimal("75000"); // 75K

        // When
        IncomeScore result = creditScoreService.calculateIncomeScore(income);

        // Proportional score calculation
        BigDecimal proportion = income.subtract(new BigDecimal("50000"))
                .divide(new BigDecimal("100000").subtract(new BigDecimal("50000")), 4, RoundingMode.HALF_EVEN);
        int expectedScore = 150 + proportion.multiply(new BigDecimal(300 - 150)).intValue();

        // Then
        assertThat(result.getScore()).isEqualTo(expectedScore);
    }

    @Test
    void shouldReturnProportionalScoreForLowIncomeRange() {
        // Given
        BigDecimal income = new BigDecimal("25000"); // 25K

        // When
        IncomeScore result = creditScoreService.calculateIncomeScore(income);

        // Proportional score calculation
        BigDecimal proportion = income.divide(new BigDecimal("50000"), 4, RoundingMode.HALF_EVEN);
        int expectedScore = proportion.multiply(new BigDecimal(150)).intValue();

        // Then
        assertThat(result.getScore()).isEqualTo(expectedScore);
    }

    @Test
    void shouldReturnLowIncomeScoreForZeroIncome() {
        // Given
        BigDecimal income = new BigDecimal("0");

        // When
        IncomeScore result = creditScoreService.calculateIncomeScore(income);

        // Then
        assertThat(result.getScore()).isEqualTo(0); // Low income score
    }

    @Test
    void shouldReturnBoundaryScoreForMediumIncomeThreshold() {
        // Given
        BigDecimal income = new BigDecimal("50000"); // Medium income threshold

        // When
        IncomeScore result = creditScoreService.calculateIncomeScore(income);

        // Then
        assertThat(result.getScore()).isEqualTo(150); // Minimum of medium range
    }

    @Test
    void shouldReturnBoundaryScoreForHighIncomeThreshold() {
        // Given
        BigDecimal income = new BigDecimal("100000"); // High income threshold

        // When
        IncomeScore result = creditScoreService.calculateIncomeScore(income);

        // Then
        assertThat(result.getScore()).isEqualTo(300); // High income score
    }

    @Test
    void shouldHandleIncomeJustBelowMediumThreshold() {
        // Given
        BigDecimal income = new BigDecimal("49999.99"); // Just below medium income threshold

        // When
        IncomeScore result = creditScoreService.calculateIncomeScore(income);

        // Proportional score calculation
        BigDecimal proportion = income.divide(new BigDecimal("50000"), 4, RoundingMode.HALF_EVEN);
        int expectedScore = proportion.multiply(new BigDecimal(150)).intValue();

        // Then
        assertThat(result.getScore()).isEqualTo(expectedScore);
    }

    @Test
    void shouldHandleIncomeJustBelowHighThreshold() {
        // Given
        BigDecimal income = new BigDecimal("99999.99"); // Just below high income threshold

        // When
        IncomeScore result = creditScoreService.calculateIncomeScore(income);

        // Proportional score calculation
        BigDecimal proportion = income.subtract(new BigDecimal("50000"))
                .divide(new BigDecimal("100000").subtract(new BigDecimal("50000")), 4, RoundingMode.HALF_EVEN);
        int expectedScore = 150 + proportion.multiply(new BigDecimal(300 - 150)).intValue();

        // Then
        assertThat(result.getScore()).isEqualTo(expectedScore);
    }
}