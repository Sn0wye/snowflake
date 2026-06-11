using FluentAssertions;
using Oxygen.Domain;
using Xunit;

namespace Oxygen.Tests.Services;

public class TermMultiplierTests
{
    [Theory]
    [InlineData(1, 1.00)]
    [InlineData(12, 1.00)]
    [InlineData(13, 1.15)]
    [InlineData(24, 1.15)]
    [InlineData(25, 1.30)]
    [InlineData(36, 1.30)]
    [InlineData(37, 1.50)]
    [InlineData(48, 1.50)]
    [InlineData(49, 1.70)]
    [InlineData(60, 1.70)]
    public void for_months_returns_correct_multiplier(int months, double expectedMultiplier)
    {
        var multiplier = TermMultiplier.For(months);

        multiplier.Value.Should().Be((decimal)expectedMultiplier);
    }
}
