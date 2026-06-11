using Oxygen.DTO;

namespace Oxygen.Service;

public interface ILoanService
{
    Task<LoanApplicationDTO> ApplyForLoan(string userId, decimal loanAmount, int term);
}
