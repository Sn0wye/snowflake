package com.snowflake.scorer.model.dto;

// Debt-to-Income Ratio Score:
// High Score = Low Ratio: 300 (Below 20%)
// Medium Score = Medium Ration: 150 (20% - 40%)
// Low Score = High Ratio: 0 (Above 40%)

public class DebtScore {
    private int score;

    public int getScore() {
        return this.score;
    }

    public void setScore(int score) {
        this.score = score;
    }

    public DebtScore(int score) {
        this.score = score;
    }

    public static DebtScore HIGH() {
        return new DebtScore(300);
    }

    public static DebtScore MEDIUM() {
        return new DebtScore(150);
    }

    public static DebtScore LOW() {
        return new DebtScore(0);
    }
}