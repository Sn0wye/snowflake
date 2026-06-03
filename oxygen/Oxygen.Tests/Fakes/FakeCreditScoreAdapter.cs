using Oxygen.Infrastructure.Adapters;

namespace Oxygen.Tests.Fakes;

public class FakeCreditScoreAdapter : ICreditScoreAdapter
{
    public int? Score { get; set; }
    public TaskCompletionSource? Gate { get; set; }
    public Action? OnCallStarted { get; set; }

    public async Task<int?> GetCreditScoreAsync(string userId)
    {
        OnCallStarted?.Invoke();
        if (Gate is not null) await Gate.Task;
        return Score;
    }
}
