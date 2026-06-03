using Oxygen.Infrastructure.Adapters;
using Pb;

namespace Oxygen.Tests.Fakes;

public class FakeCreditScoreAdapter : ICreditScoreAdapter
{
    public int? Score { get; set; }
    public TimeSpan Delay { get; set; } = TimeSpan.Zero;

    public async Task<int?> GetCreditScoreAsync(string userId)
    {
        if (Delay > TimeSpan.Zero) await Task.Delay(Delay);
        return Score;
    }
}
