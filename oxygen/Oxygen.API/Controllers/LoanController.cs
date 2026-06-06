using System.Security.Claims;
using Oxygen.API.Filters;
using Oxygen.Domain.Enums;
using Oxygen.DTO.Response;
using Oxygen.Service;
using Microsoft.AspNetCore.Mvc;
using Swashbuckle.AspNetCore.Annotations;

namespace Oxygen.API.Controllers;

[ApiController]
[Route("loan")]
[ServiceFilter(typeof(RateLimitFilter))]
public class LoanController(ILoanService loanService) : ControllerBase
{
    [SwaggerOperation(
        Summary    = "Apply for a loan",
        Description = "Apply for a loan with the desired amount and term"
    )]
    [SwaggerResponse(200, "ApplyForLoanResponse", typeof(ApplyForLoanResponse))]
    [SwaggerResponse(400, "BadRequest", typeof(ValidationErrorResponse))]
    [SwaggerResponse(401, "Unauthorized", typeof(void))]
    [HttpPost]
    public async Task<ActionResult<ApplyForLoanResponse>> ApplyForLoan([FromBody] ApplyForLoanRequest request)
    {
        var userId = User.FindFirstValue(ClaimTypes.NameIdentifier);
        
        if (userId is null) return Unauthorized();

        var application = await loanService.ApplyForLoan(
            userId,
            request.LoanAmount,
            request.Term
        );

        string? message;
        if (application.SuggestedLoan is null)
        {
            message = application.LoanApplication.Status == LoanApplicationStatus.APPROVED
                ? "Loan approved :)"
                : "Loan rejected, unfortunately we don't have a better option for you at the moment :(";

            return Ok(new ApplyForLoanResponse
            {
                Message = message,
                Status = application.LoanApplication.Status,
                Amount = application.LoanApplication.Amount,
                Term = application.LoanApplication.Term,
                SuggestedLoan = null
            });
        }

        message = application.LoanApplication.Status == LoanApplicationStatus.APPROVED
            ? "Loan approved, but we have a better option for you!"
            : "Loan rejected, but we have a better option for you!";

        return Ok(new ApplyForLoanResponse
        {
            Message = message,
            Status = application.LoanApplication.Status,
            Amount = application.LoanApplication.Amount,
            Term = application.LoanApplication.Term,
            SuggestedLoan = application.SuggestedLoan
        });
    }
}