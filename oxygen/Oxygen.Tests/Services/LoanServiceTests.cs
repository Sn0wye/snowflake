using FluentAssertions;
using Moq;
using Oxygen.Domain.Entities;
using Oxygen.Domain.Enums;
using Oxygen.Infrastructure.Adapters;
using Oxygen.Repository;
using Oxygen.Service;
using Pb;
using Xunit;

namespace Oxygen.Tests.Services;

public class LoanServiceTests
{
    private readonly Mock<ILoanRepository> _loanRepoMock;
    private readonly Mock<ICreditScoreAdapter> _creditScoreMock;
    private readonly Mock<IUsersGRPCAdapter> _usersGrpcMock;
    private readonly LoanService _sut;

    public LoanServiceTests()
    {
        _loanRepoMock = new Mock<ILoanRepository>();
        _creditScoreMock = new Mock<ICreditScoreAdapter>();
        _usersGrpcMock = new Mock<IUsersGRPCAdapter>();
        _sut = new LoanService(
            _loanRepoMock.Object,
            _creditScoreMock.Object,
            _usersGrpcMock.Object);
    }

    private static User CreateUser(string id = "user-1", long annualIncome = 100_000)
    {
        return new User
        {
            Id = id,
            Name = "Test User",
            Username = "testuser",
            Email = "test@example.com",
            AnnualIncome = annualIncome
        };
    }

    [Fact]
    public async Task apply_for_loan_approves_when_credit_score_at_least_600()
    {
        var userId = "user-1";
        var loanAmount = 10_000;
        var term = 12;
        var user = CreateUser(userId);
        _usersGrpcMock.Setup(a => a.GetUserAsync(userId)).ReturnsAsync(user);
        _creditScoreMock.Setup(a => a.GetCreditScoreAsync(userId)).ReturnsAsync(650);
        _loanRepoMock.Setup(r => r.AddAsync(It.IsAny<LoanApplication>()))
            .ReturnsAsync((LoanApplication la) => la);

        var result = await _sut.ApplyForLoan(userId, loanAmount, term);

        result.LoanApplication.Status.Should().Be(LoanApplicationStatus.APPROVED);
        result.LoanApplication.UserId.Should().Be(userId);
        result.LoanApplication.Amount.Should().Be(loanAmount);
        result.LoanApplication.Term.Should().Be(term);
    }

    [Fact]
    public async Task apply_for_loan_rejects_when_credit_score_below_600()
    {
        var userId = "user-2";
        var user = CreateUser(userId);
        _usersGrpcMock.Setup(a => a.GetUserAsync(userId)).ReturnsAsync(user);
        _creditScoreMock.Setup(a => a.GetCreditScoreAsync(userId)).ReturnsAsync(550);
        _loanRepoMock.Setup(r => r.AddAsync(It.IsAny<LoanApplication>()))
            .ReturnsAsync((LoanApplication la) => la);

        var result = await _sut.ApplyForLoan(userId, 10_000, 12);

        result.LoanApplication.Status.Should().Be(LoanApplicationStatus.REJECTED);
    }

    [Fact]
    public async Task apply_for_loan_returns_no_suggestion_when_score_is_null()
    {
        var userId = "user-3";
        var user = CreateUser(userId);
        _usersGrpcMock.Setup(a => a.GetUserAsync(userId)).ReturnsAsync(user);
        _creditScoreMock.Setup(a => a.GetCreditScoreAsync(userId)).ReturnsAsync((int?)null);
        _loanRepoMock.Setup(r => r.AddAsync(It.IsAny<LoanApplication>()))
            .ReturnsAsync((LoanApplication la) => la);

        var result = await _sut.ApplyForLoan(userId, 10_000, 12);

        result.SuggestedLoan.Should().BeNull();
        result.LoanApplication.Status.Should().Be(LoanApplicationStatus.REJECTED);
    }

    [Fact]
    public async Task apply_for_loan_persists_the_original_application()
    {
        var userId = "user-4";
        var user = CreateUser(userId);
        _usersGrpcMock.Setup(a => a.GetUserAsync(userId)).ReturnsAsync(user);
        _creditScoreMock.Setup(a => a.GetCreditScoreAsync(userId)).ReturnsAsync(700);
        _loanRepoMock.Setup(r => r.AddAsync(It.IsAny<LoanApplication>()))
            .ReturnsAsync((LoanApplication la) => la);

        await _sut.ApplyForLoan(userId, 5_000, 6);

        _loanRepoMock.Verify(r => r.AddAsync(It.Is<LoanApplication>(
            la => la.UserId == userId && la.Amount == 5_000 && la.Term == 6)), Times.Once);
    }

    [Theory]
    [InlineData(800, 0.5, 36)]
    [InlineData(900, 0.5, 36)]
    [InlineData(600, 0.35, 24)]
    [InlineData(700, 0.35, 24)]
    [InlineData(799, 0.35, 24)]
    [InlineData(300, 0.2, 12)]
    [InlineData(599, 0.2, 12)]
    public async Task apply_for_loan_suggests_better_loan_with_correct_income_and_term(
        int score, double incomeFraction, int expectedTerm)
    {
        const long annualIncome = 200_000;
        const string userId = "user-suggest";
        var user = CreateUser(userId, annualIncome);
        _usersGrpcMock.Setup(a => a.GetUserAsync(userId)).ReturnsAsync(user);
        _creditScoreMock.Setup(a => a.GetCreditScoreAsync(userId)).ReturnsAsync(score);
        _loanRepoMock.Setup(r => r.AddAsync(It.IsAny<LoanApplication>()))
            .ReturnsAsync((LoanApplication la) => la);

        var result = await _sut.ApplyForLoan(userId, 10_000, 12);

        result.SuggestedLoan.Should().NotBeNull();
        result.SuggestedLoan.Status.Should().Be(LoanApplicationStatus.APPROVED);
        result.SuggestedLoan.Amount.Should().Be(annualIncome * incomeFraction);
        result.SuggestedLoan.Term.Should().Be(expectedTerm);
        result.SuggestedLoan.UserId.Should().Be(userId);
    }

    [Fact]
    public async Task apply_for_loan_fires_user_and_score_lookups_in_parallel()
    {
        var userId = "user-5";
        _usersGrpcMock.Setup(a => a.GetUserAsync(userId))
            .ReturnsAsync(CreateUser(userId), TimeSpan.FromMilliseconds(50));
        _creditScoreMock.Setup(a => a.GetCreditScoreAsync(userId))
            .ReturnsAsync(700, TimeSpan.FromMilliseconds(50));
        _loanRepoMock.Setup(r => r.AddAsync(It.IsAny<LoanApplication>()))
            .ReturnsAsync((LoanApplication la) => la);

        var start = DateTime.UtcNow;
        await _sut.ApplyForLoan(userId, 10_000, 12);
        var elapsed = DateTime.UtcNow - start;

        elapsed.Should().BeLessThan(TimeSpan.FromMilliseconds(150));
    }
}
