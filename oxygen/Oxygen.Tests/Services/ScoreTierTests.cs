using FluentAssertions;
using Oxygen.Domain;
using Xunit;

namespace Oxygen.Tests.Services;

public class ScoreTierTests
{
    [Theory]
    [InlineData(0, "Poor", 22, 0.15)]
    [InlineData(599, "Poor", 22, 0.15)]
    [InlineData(600, "Fair", 16, 0.25)]
    [InlineData(674, "Fair", 16, 0.25)]
    [InlineData(675, "Good", 12, 0.35)]
    [InlineData(724, "Good", 12, 0.35)]
    [InlineData(725, "Very Good", 8, 0.45)]
    [InlineData(774, "Very Good", 8, 0.45)]
    [InlineData(775, "Excellent", 5, 0.55)]
    [InlineData(849, "Excellent", 5, 0.55)]
    [InlineData(850, "Outstanding", 3, 0.65)]
    [InlineData(900, "Outstanding", 3, 0.65)]
    public void for_score_returns_correct_tier(int score, string expectedName, double expectedRate, double expectedMaxLoan)
    {
        var tier = ScoreTier.For(score);

        tier.Name.Should().Be(expectedName);
        tier.BaseRate.Should().Be(expectedRate);
        tier.MaxLoanPercentage.Should().Be(expectedMaxLoan);
    }
}
