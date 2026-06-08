package com.snowflake.carbon.config;

import com.snowflake.carbon.constants.RabbitMQConstants;
import org.springframework.amqp.core.Binding;
import org.springframework.amqp.core.BindingBuilder;
import org.springframework.amqp.core.FanoutExchange;
import org.springframework.amqp.core.Queue;
import org.springframework.amqp.support.converter.MessageConverter;
import org.springframework.amqp.support.converter.SimpleMessageConverter;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class RabbitConfig {

    @Bean
    public MessageConverter messageConverter() {
        return new SimpleMessageConverter();
    }

    @Bean
    public FanoutExchange userCreatedExchange() {
        return new FanoutExchange(RabbitMQConstants.USER_CREATED_EXCHANGE, true, false);
    }

    @Bean
    public Queue calculateScoreQueue() {
        return new Queue(RabbitMQConstants.CALCULATE_SCORE_QUEUE, true);
    }

    @Bean
    public Binding calculateScoreBinding() {
        return BindingBuilder.bind(calculateScoreQueue()).to(userCreatedExchange());
    }
}
