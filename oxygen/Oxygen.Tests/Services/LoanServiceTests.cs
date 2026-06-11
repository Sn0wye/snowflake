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

    private const string DefaultUserId = "user-1";
    private const long DefaultAnnualIncome = 100_000;

    public LoanServiceTests()
    {
        _sut = new LoanService(_loanRepo, _creditScore, _usersGrpc);
        _usersGrpc.User = new User { Id = DefaultUserId, AnnualIncome = DefaultAnnualIncome };
    }

    [Fact]
    public async Task approve_loan_with_correct_rate_and_payments_when_amount_within_tier_cap()
    {
        _creditScore.Score = 650;

        var result = await _sut.ApplyForLoan(DefaultUserId, 10_000, 12);

        result.LoanApplication.Status.Should().Be(LoanApplicationStatus.APPROVED);
        result.LoanApplication.Amount.Should().Be(10_000);
        result.LoanApplication.Term.Should().Be(12);
        result.LoanApplication.InterestRate.Should().Be(16m);
        result.LoanApplication.MonthlyPayment.Should().BeApproximately(907.31m, 0.01m);
        result.LoanApplication.TotalPayment.Should().BeApproximately(10887.70m, 0.01m);
    }

    [Fact]
    public async Task reject_loan_when_requested_amount_exceeds_tier_income_cap()
    {
        _creditScore.Score = 650;

        var result = await _sut.ApplyForLoan(DefaultUserId, 50_000, 12);

        result.LoanApplication.Status.Should().Be(LoanApplicationStatus.REJECTED);
        result.RejectionReason.Should().Be("Requested amount exceeds your tier maximum.");
    }

    [Fact]
    public async Task reject_loan_when_score_is_null()
    {
        _creditScore.Score = null;

        var result = await _sut.ApplyForLoan(DefaultUserId, 10_000, 12);

        result.LoanApplication.Status.Should().Be(LoanApplicationStatus.REJECTED);
        result.RejectionReason.Should().Be("No credit score available.");
    }

    [Fact]
    public async Task persist_loan_application_with_financial_fields()
    {
        _creditScore.Score = 700;

        await _sut.ApplyForLoan(DefaultUserId, 5_000, 6);

        _loanRepo.AddedLoans.Should().ContainSingle()
            .Which.Should().Match<LoanApplication>(
                la => la.UserId == DefaultUserId
                      && la.Amount == 5_000
                      && la.Term == 6
                      && la.InterestRate > 0
                      && la.MonthlyPayment > 0
                      && la.TotalPayment > 0);
    }

    [Fact]
    public async Task apply_for_loan_starts_both_io_calls_before_awaiting_either()
    {
        var gate = new TaskCompletionSource();
        var started = new CountdownEvent(2);
        _creditScore.Gate = gate;
        _creditScore.OnCallStarted = () => started.Signal();
        _usersGrpc.Gate = gate;
        _usersGrpc.OnCallStarted = () => started.Signal();

        var sutTask = _sut.ApplyForLoan("user-5", 10_000, 12);

        started.Wait(TimeSpan.FromSeconds(5)).Should().BeTrue(
            "both adapters should start before either completes");

        gate.SetResult();
        await sutTask;
    }
}
