using Microsoft.Extensions.DependencyInjection;
using Oxygen.Infrastructure.Adapters;

namespace Oxygen.Infrastructure.Extensions;

public static class ServiceCollectionExtensions
{
    public static IServiceCollection AddCreditScoreAdapter(this IServiceCollection services)
    {
        services.AddHttpClient("CreditScorer")
            .AddStandardResilienceHandler(options =>
            {
                options.Retry.Delay = TimeSpan.FromMilliseconds(500);
                options.Retry.MaxDelay = TimeSpan.FromSeconds(5);
                options.Retry.MaxRetryAttempts = 3;
                options.CircuitBreaker.BreakDuration = TimeSpan.FromSeconds(3);
                options.CircuitBreaker.FailureRatio = 0.1;
            });

        services.AddScoped<ICreditScoreAdapter, CreditScoreAdapter>();
        return services;
    }
}
