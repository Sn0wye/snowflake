using Oxygen.DTO;
using Oxygen.Service;

namespace Oxygen.Tests.Fakes;

public class FakeLoanService : ILoanService
{
    public LoanApplicationDTO? Result { get; set; }

    public string? LastUserId { get; private set; }
    public decimal LastLoanAmount { get; private set; }
    public int LastTerm { get; private set; }

    public Task<LoanApplicationDTO> ApplyForLoan(string userId, decimal loanAmount, int term)
    {
        LastUserId = userId;
        LastLoanAmount = loanAmount;
        LastTerm = term;
        return Task.FromResult(Result ?? throw new InvalidOperationException(
            $"{nameof(FakeLoanService)}.{nameof(Result)} must be set before calling {nameof(ApplyForLoan)}."));
    }
}
