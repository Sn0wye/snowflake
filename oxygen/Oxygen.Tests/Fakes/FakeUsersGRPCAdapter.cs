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

    public TimeSpan Delay { get; set; } = TimeSpan.Zero;

    public async Task<User> GetUserAsync(string userId)
    {
        if (Delay > TimeSpan.Zero) await Task.Delay(Delay);
        return User;
    }
}
