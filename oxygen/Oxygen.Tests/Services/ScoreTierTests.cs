using FluentAssertions;
using Oxygen.Domain;
using Xunit;

namespace Oxygen.Tests.Services;

public class ScoreTierTests
{
    [Theory]
    [InlineData(0, "Poor", 0.22, 0.15)]
    [InlineData(599, "Poor", 0.22, 0.15)]
    [InlineData(600, "Fair", 0.16, 0.25)]
    [InlineData(674, "Fair", 0.16, 0.25)]
    [InlineData(675, "Good", 0.12, 0.35)]
    [InlineData(724, "Good", 0.12, 0.35)]
    [InlineData(725, "Very Good", 0.08, 0.45)]
    [InlineData(774, "Very Good", 0.08, 0.45)]
    [InlineData(775, "Excellent", 0.05, 0.55)]
    [InlineData(849, "Excellent", 0.05, 0.55)]
    [InlineData(850, "Outstanding", 0.03, 0.65)]
    [InlineData(900, "Outstanding", 0.03, 0.65)]
    public void for_score_returns_correct_tier(int score, string expectedName, double expectedRate, double expectedMaxLoanFraction)
    {
        var tier = ScoreTier.For(score);

        tier.Name.Should().Be(expectedName);
        tier.Rate.Should().Be((decimal)expectedRate);
        tier.MaxLoanFraction.Should().Be((decimal)expectedMaxLoanFraction);
    }
}
