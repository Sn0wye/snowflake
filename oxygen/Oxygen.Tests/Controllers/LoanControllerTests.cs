using System.Security.Claims;
using FluentAssertions;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Oxygen.API.Controllers;
using Oxygen.Domain.Entities;
using Oxygen.Domain.Enums;
using Oxygen.DTO;
using Oxygen.DTO.Response;
using Oxygen.Tests.Fakes;
using Xunit;

namespace Oxygen.Tests.Controllers;

public class LoanControllerTests
{
    private readonly FakeLoanService _loanService = new();
    private readonly LoanController _sut;

    public LoanControllerTests()
    {
        _sut = new LoanController(_loanService);
    }

    private void SetUserIdClaim(string userId)
    {
        var user = new ClaimsPrincipal(new ClaimsIdentity([
            new Claim(ClaimTypes.NameIdentifier, userId)
        ]));
        _sut.ControllerContext = new ControllerContext
        {
            HttpContext = new DefaultHttpContext { User = user }
        };
    }

    private static LoanApplicationDTO CreateApplication(
        LoanApplicationStatus status,
        decimal interestRate = 5m,
        decimal monthlyPayment = 100m,
        decimal totalPayment = 1200m,
        string? rejectionReason = null)
    {
        return new LoanApplicationDTO
        {
            LoanApplication = new LoanApplication
            {
                UserId = "user-1",
                Status = status,
                Amount = 10_000,
                Term = 12,
                InterestRate = interestRate,
                MonthlyPayment = monthlyPayment,
                TotalPayment = totalPayment
            },
            RejectionReason = rejectionReason
        };
    }

    [Fact]
    public async Task apply_for_loan_returns_401_when_user_id_claim_is_missing()
    {
        _sut.ControllerContext = new ControllerContext
        {
            HttpContext = new DefaultHttpContext()
        };

        var result = await _sut.ApplyForLoan(new ApplyForLoanRequest { LoanAmount = 10_000, Term = 12 });

        result.Result.Should().BeOfType<UnauthorizedResult>();
    }

    [Fact]
    public async Task apply_for_loan_returns_approved_message_when_status_is_approved()
    {
        SetUserIdClaim("user-1");
        _loanService.Result = CreateApplication(LoanApplicationStatus.APPROVED);

        var result = await _sut.ApplyForLoan(new ApplyForLoanRequest { LoanAmount = 10_000, Term = 12 });

        var okResult = result.Result.Should().BeOfType<OkObjectResult>().Subject;
        var response = okResult.Value.Should().BeOfType<ApplyForLoanResponse>().Subject;
        response.Status.Should().Be(LoanApplicationStatus.APPROVED);
        response.Message.Should().Be("Loan approved :)");
    }

    [Fact]
    public async Task apply_for_loan_returns_rejected_message_when_status_is_rejected()
    {
        SetUserIdClaim("user-1");
        _loanService.Result = CreateApplication(LoanApplicationStatus.REJECTED,
            rejectionReason: "Requested amount exceeds your tier maximum.");

        var result = await _sut.ApplyForLoan(new ApplyForLoanRequest { LoanAmount = 10_000, Term = 12 });

        var okResult = result.Result.Should().BeOfType<OkObjectResult>().Subject;
        var response = okResult.Value.Should().BeOfType<ApplyForLoanResponse>().Subject;
        response.Status.Should().Be(LoanApplicationStatus.REJECTED);
        response.Message.Should().Be("Loan rejected: Requested amount exceeds your tier maximum.");
    }

    [Fact]
    public async Task apply_for_loan_returns_financial_fields_from_application()
    {
        SetUserIdClaim("user-1");
        _loanService.Result = CreateApplication(LoanApplicationStatus.APPROVED, 16m, 907.31m, 10887.70m);

        var result = await _sut.ApplyForLoan(new ApplyForLoanRequest { LoanAmount = 10_000, Term = 12 });

        var okResult = result.Result.Should().BeOfType<OkObjectResult>().Subject;
        var response = okResult.Value.Should().BeOfType<ApplyForLoanResponse>().Subject;
        response.InterestRate.Should().Be(16m);
        response.MonthlyPayment.Should().Be(907.31m);
        response.TotalPayment.Should().Be(10887.70m);
        response.Amount.Should().Be(10_000);
        response.Term.Should().Be(12);
    }
}
