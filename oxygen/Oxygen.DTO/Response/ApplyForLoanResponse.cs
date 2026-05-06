using Oxygen.Domain.Enums;

namespace Oxygen.DTO.Response;

public class ApplyForLoanResponse
{
    // <summary> Loan application status </summary>
    // <example>APPROVED</example>
    public LoanApplicationStatus Status { get; set; }
    
    // <summary> Message </summary>
    // <example>Loan approved :)</example>
    public required string Message { get; set; }
    
    // <summary> Loan amount </summary>
    // <example>1000.00</example>
    public double Amount { get; set; }
    
    // <summary> Loan term in months</summary>
    // <example>12</example>
    public int Term { get; set; }
    
    // <summary> Suggested loan application </summary>
    // <example>null</example>
    public Oxygen.Domain.Entities.LoanApplication? SuggestedLoan { get; set; }
}