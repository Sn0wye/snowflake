namespace Oxygen.Infrastructure.Adapters;

public class CreditScoreAdapter: ICreditScoreAdapter
{
    private readonly HttpClient _httpClient = new();
    
    public async Task<int?> GetCreditScoreAsync(string userId)
    {
        var response = await _httpClient.GetAsync($"http://localhost/scorer/{userId}");
        if (!response.IsSuccessStatusCode) return null;
        
        var content = await response.Content.ReadAsStringAsync();
        return int.Parse(content);
    }
}

public interface ICreditScoreAdapter
{
    Task<int?> GetCreditScoreAsync(string userId);
}