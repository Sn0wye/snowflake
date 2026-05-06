namespace Oxygen.DTO;

public class LoanApplicationDTO
{
    public required Domain.Entities.LoanApplication LoanApplication { get; init; }
    public Domain.Entities.LoanApplication? SuggestedLoan { get; init; }
}