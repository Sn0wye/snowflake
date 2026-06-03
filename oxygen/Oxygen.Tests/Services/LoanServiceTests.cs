using System.Threading;
using FluentAssertions;
using Oxygen.Domain.Entities;
using Oxygen.Domain.Enums;
using Oxygen.Service;
using Oxygen.Tests.Fakes;
using Pb;
using Xunit;

namespace Oxygen.Tests.Services;

public class LoanServiceTests
{
    private readonly FakeLoanRepository _loanRepo = new();
    private readonly FakeCreditScoreAdapter _creditScore = new();
    private readonly FakeUsersGRPCAdapter _usersGrpc = new();
    private readonly LoanService _sut;

    public LoanServiceTests()
    {
        _sut = new LoanService(_loanRepo, _creditScore, _usersGrpc);
    }

    [Fact]
    public async Task apply_for_loan_approves_when_credit_score_at_least_600()
    {
        const string userId = "user-1";
        var user = new User { Id = userId, AnnualIncome = 100_000 };
        _usersGrpc.User = user;
        _creditScore.Score = 650;

        var result = await _sut.ApplyForLoan(userId, 10_000, 12);

        result.LoanApplication.Status.Should().Be(LoanApplicationStatus.APPROVED);
        result.LoanApplication.UserId.Should().Be(userId);
        result.LoanApplication.Amount.Should().Be(10_000);
        result.LoanApplication.Term.Should().Be(12);
    }

    [Fact]
    public async Task apply_for_loan_rejects_when_credit_score_below_600()
    {
        const string userId = "user-2";
        _usersGrpc.User = new User { Id = userId, AnnualIncome = 100_000 };
        _creditScore.Score = 550;

        var result = await _sut.ApplyForLoan(userId, 10_000, 12);

        result.LoanApplication.Status.Should().Be(LoanApplicationStatus.REJECTED);
    }

    [Fact]
    public async Task apply_for_loan_returns_no_suggestion_when_score_is_null()
    {
        const string userId = "user-3";
        _usersGrpc.User = new User { Id = userId, AnnualIncome = 100_000 };
        _creditScore.Score = null;

        var result = await _sut.ApplyForLoan(userId, 10_000, 12);

        result.SuggestedLoan.Should().BeNull();
        result.LoanApplication.Status.Should().Be(LoanApplicationStatus.REJECTED);
    }

    [Fact]
    public async Task apply_for_loan_persists_the_original_application()
    {
        const string userId = "user-4";
        _usersGrpc.User = new User { Id = userId, AnnualIncome = 100_000 };
        _creditScore.Score = 700;

        await _sut.ApplyForLoan(userId, 5_000, 6);

        _loanRepo.AddedLoans.Should().ContainSingle()
            .Which.Should().Match<LoanApplication>(
                la => la.UserId == userId && la.Amount == 5_000 && la.Term == 6);
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
        _usersGrpc.User = new User { Id = userId, AnnualIncome = annualIncome };
        _creditScore.Score = score;

        var result = await _sut.ApplyForLoan(userId, 10_000, 12);

        result.SuggestedLoan.Should().NotBeNull();
        result.SuggestedLoan!.Status.Should().Be(LoanApplicationStatus.APPROVED);
        result.SuggestedLoan.Amount.Should().Be(annualIncome * incomeFraction);
        result.SuggestedLoan.Term.Should().Be(expectedTerm);
        result.SuggestedLoan.UserId.Should().Be(userId);
    }

    [Fact]
    public async Task apply_for_loan_starts_both_io_calls_before_awaiting_either()
    {
        var gate = new TaskCompletionSource();
        var callCount = 0;
        _creditScore.Gate = gate;
        _creditScore.OnCallStarted = () => Interlocked.Increment(ref callCount);
        _usersGrpc.Gate = gate;
        _usersGrpc.OnCallStarted = () => Interlocked.Increment(ref callCount);

        var sutTask = _sut.ApplyForLoan("user-5", 10_000, 12);

        // Both calls should have started (blocked on gate) while we're still waiting
        await Task.Delay(50);
        Interlocked.CompareExchange(ref callCount, 0, 0).Should().Be(2);

        gate.SetResult();
        await sutTask;
    }
}
