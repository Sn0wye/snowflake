package com.snowflake.carbon.config;

import com.snowflake.carbon.constants.RabbitMQConstants;
import org.springframework.amqp.core.Binding;
import org.springframework.amqp.core.BindingBuilder;
import org.springframework.amqp.core.FanoutExchange;
import org.springframework.amqp.core.Queue;
import org.springframework.amqp.rabbit.connection.ConnectionFactory;
import org.springframework.amqp.rabbit.connection.CachingConnectionFactory;
import org.springframework.amqp.support.converter.MessageConverter;
import org.springframework.amqp.support.converter.SimpleMessageConverter;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class RabbitConfig {

    @Value("${spring.rabbitmq.host}")
    private String host;

    @Value("${spring.rabbitmq.port}")
    private int port;

    @Value("${spring.rabbitmq.addresses:#{null}}")
    private String addresses;

    @Value("${spring.rabbitmq.username}")
    private String username;

    @Value("${spring.rabbitmq.password}")
    private String password;

    @Bean
    public ConnectionFactory connectionFactory() {
        CachingConnectionFactory connectionFactory;
        if (addresses != null && !addresses.isBlank()) {
            connectionFactory = new CachingConnectionFactory();
            connectionFactory.setAddresses(addresses);
        } else {
            connectionFactory = new CachingConnectionFactory(host, port);
        }
        connectionFactory.setUsername(username);
        connectionFactory.setPassword(password);
        return connectionFactory;
    }

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
