using System.ComponentModel.DataAnnotations;

public class ApplyForLoanRequest
{
    [Required(ErrorMessage = "Loan amount is required.")]
    [Range(300, double.MaxValue, ErrorMessage = "Loan amount must be greater than 300.")]
    public double LoanAmount { get; set; }

    [Required(ErrorMessage = "Term is required.")]
    [Range(1, 60, ErrorMessage = "Term must be between 1 and 60.")]
    public int Term { get; set; }
}