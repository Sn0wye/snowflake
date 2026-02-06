package com.snowflake.scorer.consumers;

public class ScoreUpdatedEvent {
    private String userId;
    private int creditScore;

    public ScoreUpdatedEvent(String userId, int creditScore) {
        this.userId = userId;
        this.creditScore = creditScore;
    }

    public String getUserId() {
        return userId;
    }

    public void setUserId(String userId) {
        this.userId = userId;
    }

    public int getCreditScore() {
        return creditScore;
    }

    public void setCreditScore(int creditScore) {
        this.creditScore = creditScore;
    }

    @Override
    public String toString() {
        return "ScoreUpdatedEvent{" +
                "scoreId='" + userId + '\'' +
                ", creditScore=" + creditScore +
                '}';
    }
}
