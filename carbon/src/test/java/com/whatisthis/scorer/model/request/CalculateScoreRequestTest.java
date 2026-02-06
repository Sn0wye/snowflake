package com.snowflake.scorer.model.request;

import jakarta.validation.ConstraintViolation;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import jakarta.validation.Validation;
import jakarta.validation.Validator;
import jakarta.validation.ValidatorFactory;

import java.util.Set;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

import static org.junit.jupiter.api.Assertions.assertDoesNotThrow;

class CalculateScoreRequestTest {

    private Validator validator;

    @BeforeEach
    void setUp() {
        ValidatorFactory factory = Validation.buildDefaultValidatorFactory();
        validator = factory.getValidator();
    }

    @Test
    void validRequest_shouldPassValidation() {
        // Arrange
        assertDoesNotThrow(() -> new CalculateScoreRequest(
                10000.0,   // $100.00
                5000.0,    // $50.00
                20000.0    // $200.00
        ));

        CalculateScoreRequest request = new CalculateScoreRequest(
                10000.0,   // $100.00
                5000.0,    // $50.00
                20000.0    // $200.00
        );

        // Act
        Set<ConstraintViolation<CalculateScoreRequest>> violations = validator.validate(request);

        // Assert
        assertTrue(violations.isEmpty(), "Request should be valid");
    }

    @Test
    void zeroIncome_shouldFailValidation() {
        // Arrange & Act
        Set<ConstraintViolation<CalculateScoreRequest>> violations = validator.validate(
                new CalculateScoreRequest(
                        0.0,
                        5000.0,
                        20000.0
                )
        );

        // Assert
        assertEquals(1, violations.size(), "Should have one violation");
        assertTrue(violations.stream()
                        .anyMatch(v -> "Income must be greater than zero.".equals(v.getMessage())),
                "Violation message should match");
    }

    @Test
    void negativeDebt_shouldFailValidation() {
        // Arrange & Act
        Set<ConstraintViolation<CalculateScoreRequest>> violations = validator.validate(
                new CalculateScoreRequest(
                        10000.0,
                        -1.0,
                        20000.0
                )
        );

        // Assert
        assertEquals(1, violations.size(), "Should have one violation");
        assertTrue(violations.stream()
                        .anyMatch(v -> "Debt cannot be negative.".equals(v.getMessage())),
                "Violation message should match");
    }

    @Test
    void negativeAssetsValue_shouldFailValidation() {
        // Arrange & Act
        Set<ConstraintViolation<CalculateScoreRequest>> violations = validator.validate(
                new CalculateScoreRequest(
                        10000.0,
                        5000.0,
                        -1.0
                )
        );

        // Assert
        assertEquals(1, violations.size(), "Should have one violation");
        assertTrue(violations.stream()
                        .anyMatch(v -> "Assets value cannot be negative.".equals(v.getMessage())),
                "Violation message should match");
    }
}