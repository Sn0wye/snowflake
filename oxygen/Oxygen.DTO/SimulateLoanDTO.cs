namespace Oxygen.DTO;

public class SimulateLoanDTO
{
    public required int Id { get; set; }
    public required int ApprovedAmount { get; set; }
    public required decimal InterestRate { get; set; }
    public required decimal MonthlyPayment { get; set; }
    public required decimal TotalPayment { get; set; }
    public required int Term { get; set; }
}