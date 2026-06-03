using Oxygen.Infrastructure.Adapters;
using Pb;

namespace Oxygen.Tests.Fakes;

public class FakeUsersGRPCAdapter : IUsersGRPCAdapter
{
    public User User { get; set; } = new()
    {
        Id = "user-1",
        Name = "Test User",
        Username = "testuser",
        Email = "test@example.com",
        AnnualIncome = 100_000
    };

    public TaskCompletionSource? Gate { get; set; }
    public Action? OnCallStarted { get; set; }

    public async Task<User> GetUserAsync(string userId)
    {
        OnCallStarted?.Invoke();
        if (Gate is not null) await Gate.Task;
        return User;
    }
}
