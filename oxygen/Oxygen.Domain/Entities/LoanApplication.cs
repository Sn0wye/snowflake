using System.ComponentModel.DataAnnotations.Schema;
using Oxygen.Domain.Enums;

namespace Oxygen.Domain.Entities;

[Table("loan_applications")]
public class LoanApplication
{
    public int Id { get; set; }
    public required string UserId { get; set; }
    public LoanApplicationStatus Status { get; set; } = LoanApplicationStatus.PENDING;
    public decimal Amount { get; set; }
    public int Term { get; set; }
    public decimal? InterestRate { get; set; }
    public decimal? MonthlyPayment { get; set; }
    public decimal? TotalPayment { get; set; }

    public void ChangeStatus(LoanApplicationStatus applicationStatus)
    {
        Status = applicationStatus;
    }
}
