using Oxygen.Domain.Entities;
using Oxygen.Repository;

namespace Oxygen.Tests.Fakes;

public class FakeLoanRepository : ILoanRepository
{
    public List<LoanApplication> AddedLoans { get; } = [];
    public LoanApplication? FindResult { get; set; }

    public Task<LoanApplication?> FindAsync(int id)
    {
        return Task.FromResult(FindResult);
    }

    public Task<LoanApplication> AddAsync(LoanApplication loanApplication)
    {
        AddedLoans.Add(loanApplication);
        return Task.FromResult(loanApplication);
    }
}
