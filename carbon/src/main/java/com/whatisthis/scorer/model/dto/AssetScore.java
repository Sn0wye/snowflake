package com.snowflake.scorer.model.dto;

// Assets Ownership Score:
// High value assets: 300 (Value above $500,000)
// Medium value assets: 150 (Value between $200,000 and $500,000)
// Low value assets / Nothing: 0 (Below $200,000)
public class AssetScore {
    private int score;

    public int getScore() {
        return this.score;
    }

    public void setScore(int score) {
        this.score = score;
    }

    public AssetScore(int score) {
        this.score = score;
    }

    public static AssetScore HIGH() {
        return new AssetScore(300);
    }

    public static AssetScore MEDIUM() {
        return new AssetScore(150);
    }

    public static AssetScore LOW() {
        return new AssetScore(0);
    }
}
