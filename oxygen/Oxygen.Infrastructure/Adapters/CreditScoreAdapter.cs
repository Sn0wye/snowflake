using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Configuration;

namespace Oxygen.Infrastructure.Adapters;

public class CreditScoreAdapter : ICreditScoreAdapter
{
    private readonly IHttpClientFactory _httpClientFactory;
    private readonly IHttpContextAccessor _httpContextAccessor;
    private readonly string _scorerBaseUrl;

    public CreditScoreAdapter(IConfiguration configuration, IHttpClientFactory httpClientFactory, IHttpContextAccessor httpContextAccessor)
    {
        _scorerBaseUrl = configuration.GetConnectionString("CreditScorer")
            ?? throw new InvalidOperationException(
                "Configuration 'ConnectionStrings:CreditScorer' is not set or is empty.");
        _httpClientFactory = httpClientFactory;
        _httpContextAccessor = httpContextAccessor;
    }

    public async Task<int?> GetCreditScoreAsync(string userId)
    {
        var client = _httpClientFactory.CreateClient("CreditScorer");
        var request = new HttpRequestMessage(HttpMethod.Get,
            $"{_scorerBaseUrl.TrimEnd('/')}/{Uri.EscapeDataString(userId)}");

        var correlationId = _httpContextAccessor.HttpContext?.Items["correlation_id"] as string;
        if (!string.IsNullOrEmpty(correlationId))
        {
            request.Headers.Add("X-Correlation-ID", correlationId);
        }

        using var response = await client.SendAsync(request);
        if (!response.IsSuccessStatusCode) return null;

        var content = await response.Content.ReadAsStringAsync();
        return int.TryParse(content, out var score) ? score : null;
    }
}

public interface ICreditScoreAdapter
{
    Task<int?> GetCreditScoreAsync(string userId);
}