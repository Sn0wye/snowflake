package com.snowflake.scorer.repositories;

import org.springframework.data.jpa.repository.JpaRepository;
import com.snowflake.scorer.entities.Score;

import java.util.Optional;

public interface ScoreRepository extends JpaRepository<Score, String> {
    Optional<Score> findByUserId(String userId);
}