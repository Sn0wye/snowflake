package com.snowflake.carbon;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.system.CapturedOutput;
import org.springframework.boot.test.system.OutputCaptureExtension;
import org.springframework.test.context.ActiveProfiles;

// Boots the production profile so the LogstashEncoder (JSON) appender is active,
// then verifies the static service field lands on a real log line.
@SpringBootTest
@ActiveProfiles("production")
@ExtendWith(OutputCaptureExtension.class)
class ServiceLogFieldTest {

    private static final Logger log = LoggerFactory.getLogger(ServiceLogFieldTest.class);

    @Test
    void jsonLogLineIncludesServiceField(CapturedOutput output) {
        log.info("service-field-probe");

        assertThat(output).contains("\"service\":\"carbon\"");
    }
}
