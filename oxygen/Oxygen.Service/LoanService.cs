using Oxygen.Domain.Enums;
using Oxygen.DTO;
using Oxygen.Infrastructure.Adapters;
using Oxygen.Repository;
using Pb;

namespace Oxygen.Service;

public class LoanService(
    ILoanRepository loanRepository,
    ICreditScoreAdapter creditScoreAdapter,
    IUsersGRPCAdapter usersGrpcAdapter)
    : ILoanService
{
    public async Task<LoanApplicationDTO> ApplyForLoan(string userId, double loanAmount, int term)
    {
        var userTask = usersGrpcAdapter.GetUserAsync(userId);
        var scoreTask = creditScoreAdapter.GetCreditScoreAsync(userId);

        var user = await userTask;
        var score = await scoreTask;

        if (score is null)
            return await RejectAsync(userId, loanAmount, term);

        var tier = Domain.ScoreTier.For(score.Value);
        var maxAmount = user.AnnualIncome * tier.MaxLoanPercentage;

        if (loanAmount > maxAmount)
            return await RejectAsync(userId, loanAmount, term);

        var termMultiplier = Domain.TermMultiplier.For(term);
        var finalRatePercent = tier.BaseRate * termMultiplier.Value;
        var monthlyRate = finalRatePercent / 12 / 100;
        var factor = Math.Pow(1 + monthlyRate, term);
        var monthlyPayment = loanAmount * monthlyRate * factor / (factor - 1);
        var totalPayment = monthlyPayment * term;

        var loan = new Domain.Entities.LoanApplication
        {
            UserId = userId,
            Amount = loanAmount,
            Term = term,
            Status = LoanApplicationStatus.APPROVED,
            InterestRate = (decimal)finalRatePercent,
            MonthlyPayment = (decimal)monthlyPayment,
            TotalPayment = (decimal)totalPayment
        };

        await loanRepository.AddAsync(loan);

        return new LoanApplicationDTO { LoanApplication = loan };
    }

    private async Task<LoanApplicationDTO> RejectAsync(string userId, double amount, int term)
    {
        var loan = new Domain.Entities.LoanApplication
        {
            UserId = userId,
            Amount = amount,
            Term = term,
            Status = LoanApplicationStatus.REJECTED
        };

        await loanRepository.AddAsync(loan);

        return new LoanApplicationDTO { LoanApplication = loan };
    }
}
