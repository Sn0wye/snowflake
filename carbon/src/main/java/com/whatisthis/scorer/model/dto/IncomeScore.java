package com.snowflake.scorer.model.dto;

// Income Score:
// High Income: 300 (Above $100,000 per year)
// Medium Income: 150 ($50,000 - $100,000 per year)
// Low Income: 0 (Below $50,000 per year)

public class IncomeScore {
    private int score;

    public int getScore() {
        return this.score;
    }

    public void setScore(int score) {
        this.score = score;
    }

    public IncomeScore(int score) {
        this.score = score;
    }

    public static IncomeScore HIGH() {
        return new IncomeScore(300);
    }

    public static IncomeScore MEDIUM() {
        return new IncomeScore(150);
    }

    public static IncomeScore LOW() {
        return new IncomeScore(0);
    }
}
