package com.snowflake.scorer.services;

import com.snowflake.scorer.entities.Score;
import com.snowflake.scorer.exceptions.ScoreNotFoundException;
import com.snowflake.scorer.repositories.ScoreRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;

@Service
public class ScoreService {
    @Autowired
    private CreditScoreService creditScoreService;

    @Autowired
    private ScoreRepository scoreRepository;

    public Score calculateScore(String userId, BigDecimal income, BigDecimal debt, BigDecimal assetsValue) {
        int creditScore = creditScoreService.calculateCreditScore(income, debt, assetsValue);

        Score score = scoreRepository.findByUserId(userId).orElseGet(Score::new);

        score.setUserId(userId);
        score.setIncome(income);
        score.setDebt(debt);
        score.setAssetsValue(assetsValue);
        score.setCreditScore(creditScore);

        scoreRepository.save(score);

        return score;
    }

    public Score getScore(String userId) {
        return scoreRepository.findByUserId(userId).orElseThrow(() -> new ScoreNotFoundException("Score not found"));
    }
}
