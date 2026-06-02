package com.snowflake.carbon.repositories;

import org.springframework.data.jpa.repository.JpaRepository;
import com.snowflake.carbon.entities.Score;

import java.util.Optional;
import java.util.UUID;

public interface ScoreRepository extends JpaRepository<Score, UUID> {
    Optional<Score> findByUserId(String userId);
}