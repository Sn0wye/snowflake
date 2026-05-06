using Oxygen.DTO;

namespace Oxygen.Service;

public interface ILoanService
{
    Task<LoanApplicationDTO> ApplyForLoan(string userId, double loanAmount, int term);
}