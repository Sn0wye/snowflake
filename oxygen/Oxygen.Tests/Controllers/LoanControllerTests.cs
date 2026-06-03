using System.Security.Claims;
using FluentAssertions;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Moq;
using Oxygen.API.Controllers;
using Oxygen.Domain.Entities;
using Oxygen.Domain.Enums;
using Oxygen.DTO;
using Oxygen.DTO.Response;
using Oxygen.Service;
using Xunit;

namespace Oxygen.Tests.Controllers;

public class LoanControllerTests
{
    private readonly Mock<ILoanService> _loanServiceMock;
    private readonly LoanController _sut;

    public LoanControllerTests()
    {
        _loanServiceMock = new Mock<ILoanService>();
        _sut = new LoanController(_loanServiceMock.Object);
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
        LoanApplicationStatus status, LoanApplication? suggestedLoan = null)
    {
        return new LoanApplicationDTO
        {
            LoanApplication = new LoanApplication
            {
                UserId = "user-1",
                Status = status,
                Amount = 10_000,
                Term = 12
            },
            SuggestedLoan = suggestedLoan
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
    public async Task apply_for_loan_returns_approved_message_when_status_is_approved_and_no_suggestion()
    {
        SetUserIdClaim("user-1");
        var dto = CreateApplication(LoanApplicationStatus.APPROVED);
        _loanServiceMock.Setup(s => s.ApplyForLoan("user-1", 10_000, 12))
            .ReturnsAsync(dto);

        var result = await _sut.ApplyForLoan(new ApplyForLoanRequest { LoanAmount = 10_000, Term = 12 });

        var okResult = result.Result.Should().BeOfType<OkObjectResult>().Subject;
        var response = okResult.Value.Should().BeOfType<ApplyForLoanResponse>().Subject;
        response.Status.Should().Be(LoanApplicationStatus.APPROVED);
        response.Message.Should().Be("Loan approved :)");
        response.SuggestedLoan.Should().BeNull();
    }

    [Fact]
    public async Task apply_for_loan_returns_rejected_message_when_status_is_rejected_and_no_suggestion()
    {
        SetUserIdClaim("user-1");
        var dto = CreateApplication(LoanApplicationStatus.REJECTED);
        _loanServiceMock.Setup(s => s.ApplyForLoan("user-1", 10_000, 12))
            .ReturnsAsync(dto);

        var result = await _sut.ApplyForLoan(new ApplyForLoanRequest { LoanAmount = 10_000, Term = 12 });

        var okResult = result.Result.Should().BeOfType<OkObjectResult>().Subject;
        var response = okResult.Value.Should().BeOfType<ApplyForLoanResponse>().Subject;
        response.Status.Should().Be(LoanApplicationStatus.REJECTED);
        response.Message.Should().Contain("rejected");
    }

    [Fact]
    public async Task apply_for_loan_returns_better_option_message_when_approved_with_suggestion()
    {
        SetUserIdClaim("user-1");
        var suggested = new LoanApplication
        {
            UserId = "user-1",
            Status = LoanApplicationStatus.APPROVED,
            Amount = 50_000,
            Term = 36
        };
        var dto = CreateApplication(LoanApplicationStatus.APPROVED, suggested);
        _loanServiceMock.Setup(s => s.ApplyForLoan("user-1", 10_000, 12))
            .ReturnsAsync(dto);

        var result = await _sut.ApplyForLoan(new ApplyForLoanRequest { LoanAmount = 10_000, Term = 12 });

        var okResult = result.Result.Should().BeOfType<OkObjectResult>().Subject;
        var response = okResult.Value.Should().BeOfType<ApplyForLoanResponse>().Subject;
        response.Status.Should().Be(LoanApplicationStatus.APPROVED);
        response.Message.Should().Be("Loan approved, but we have a better option for you!");
        response.SuggestedLoan.Should().NotBeNull();
        response.SuggestedLoan!.Amount.Should().Be(50_000);
    }

    [Fact]
    public async Task apply_for_loan_returns_better_option_message_when_rejected_with_suggestion()
    {
        SetUserIdClaim("user-1");
        var suggested = new LoanApplication
        {
            UserId = "user-1",
            Status = LoanApplicationStatus.APPROVED,
            Amount = 20_000,
            Term = 24
        };
        var dto = CreateApplication(LoanApplicationStatus.REJECTED, suggested);
        _loanServiceMock.Setup(s => s.ApplyForLoan("user-1", 10_000, 12))
            .ReturnsAsync(dto);

        var result = await _sut.ApplyForLoan(new ApplyForLoanRequest { LoanAmount = 10_000, Term = 12 });

        var okResult = result.Result.Should().BeOfType<OkObjectResult>().Subject;
        var response = okResult.Value.Should().BeOfType<ApplyForLoanResponse>().Subject;
        response.Status.Should().Be(LoanApplicationStatus.REJECTED);
        response.Message.Should().Be("Loan rejected, but we have a better option for you!");
        response.SuggestedLoan.Should().NotBeNull();
    }
}
