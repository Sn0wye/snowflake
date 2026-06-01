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

        var loan = new Domain.Entities.LoanApplication
        {
            UserId = userId,
            Amount = loanAmount,
            Term = term,
            Status = score >= 600 ? LoanApplicationStatus.APPROVED : LoanApplicationStatus.REJECTED
        };

        await loanRepository.AddAsync(loan);

        var suggestedLoan = score.HasValue 
            ? await SuggestBetterLoan(user, score.Value) 
            : null;

        return new LoanApplicationDTO
        {
            LoanApplication = loan,
            SuggestedLoan = suggestedLoan
        };
    }

    private async Task<Domain.Entities.LoanApplication?> SuggestBetterLoan(User user, int score)
    {
        double incomePercentage = score switch
        {
            >= 800 => 0.5,
            < 600 => 0.2,
            _ => 0.35
        };
        int term = score switch
        {
            >= 800 => 36,
            < 600 => 12,
            _ => 24
        };

        var suggestedLoan = new Domain.Entities.LoanApplication
        {
            UserId = user.Id,
            Amount = user.AnnualIncome * incomePercentage,
            Term = term,
            Status = LoanApplicationStatus.APPROVED
        };

        return suggestedLoan;
    }
}