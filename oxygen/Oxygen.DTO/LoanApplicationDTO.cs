namespace Oxygen.DTO;

public class LoanApplicationDTO
{
    public required Domain.Entities.LoanApplication LoanApplication { get; init; }
    public string? RejectionReason { get; init; }
}