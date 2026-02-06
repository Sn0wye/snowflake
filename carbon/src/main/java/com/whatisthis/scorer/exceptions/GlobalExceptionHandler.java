package com.snowflake.scorer.exceptions;

import com.snowflake.scorer.model.response.ErrorResponse;
import com.snowflake.scorer.model.response.ScoreNotFoundResponse;
import com.snowflake.scorer.model.response.ValidationErrorResponse;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;

import java.util.HashMap;
import java.util.Map;

@ControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(RuntimeException.class)
    public ResponseEntity<ErrorResponse> handleRuntimeException(RuntimeException ex) {
        ErrorResponse body = new ErrorResponse(ex.getMessage(), HttpStatus.INTERNAL_SERVER_ERROR.value());
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(body);
    }

    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<ValidationErrorResponse> handleValidationExceptions(MethodArgumentNotValidException ex) {
        Map<String, String> errors = new HashMap<>();
        ex.getBindingResult().getFieldErrors().forEach(error ->
                errors.put(error.getField(), error.getDefaultMessage())
        );
        ValidationErrorResponse body = new ValidationErrorResponse("Invalid request body", HttpStatus.BAD_REQUEST.value(), errors);

        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(body);
    }

    @ExceptionHandler(ScoreNotFoundException.class)
    public ResponseEntity<ScoreNotFoundResponse> handleScoreNotFoundException(ScoreNotFoundException ex) {
        ScoreNotFoundResponse body = new ScoreNotFoundResponse(ex.getMessage(), HttpStatus.NOT_FOUND.value());

        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(body);
    }

    @ExceptionHandler(UnauthorizedException.class)
    public ResponseEntity<Map<String, String>> handleUnauthorizedException(UnauthorizedException ex) {
        Map<String, String> body = new HashMap<>();
        body.put("message", ex.getMessage());
        body.put("status_code", String.valueOf(HttpStatus.UNAUTHORIZED.value()));
        return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(body);
    }
}
