using Microsoft.Extensions.Configuration;

namespace Oxygen.Infrastructure.Adapters;

public class CreditScoreAdapter : ICreditScoreAdapter
{
    private readonly IHttpClientFactory _httpClientFactory;
    private readonly string _scorerBaseUrl;

    public CreditScoreAdapter(IConfiguration configuration, IHttpClientFactory httpClientFactory)
    {
        _scorerBaseUrl = configuration.GetConnectionString("CreditScorer");
        if (string.IsNullOrWhiteSpace(_scorerBaseUrl))
        {
            throw new InvalidOperationException(
                "Configuration 'ConnectionStrings:CreditScorer' is not set or is empty.");
        }
        _httpClientFactory = httpClientFactory;
    }

    public async Task<int?> GetCreditScoreAsync(string userId)
    {
        var client = _httpClientFactory.CreateClient();
        var response = await client.GetAsync($"{_scorerBaseUrl.TrimEnd('/')}/{userId}");
        if (!response.IsSuccessStatusCode) return null;

        var content = await response.Content.ReadAsStringAsync();
        return int.Parse(content);
    }
}

public interface ICreditScoreAdapter
{
    Task<int?> GetCreditScoreAsync(string userId);
}