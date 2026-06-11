namespace Oxygen.Domain;

public class ScoreTier
{
    public string Name { get; }
    public decimal BaseRate { get; }
    public decimal MaxLoanPercentage { get; }

    private ScoreTier(string name, decimal baseRate, decimal maxLoanPercentage)
    {
        Name = name;
        BaseRate = baseRate;
        MaxLoanPercentage = maxLoanPercentage;
    }

    public static ScoreTier For(int score)
    {
        if (score <= 599)
            return new ScoreTier("Poor", 22m, 0.15m);

        return score switch
        {
            <= 674 => new ScoreTier("Fair", 16m, 0.25m),
            <= 724 => new ScoreTier("Good", 12m, 0.35m),
            <= 774 => new ScoreTier("Very Good", 8m, 0.45m),
            <= 849 => new ScoreTier("Excellent", 5m, 0.55m),
            _ => new ScoreTier("Outstanding", 3m, 0.65m)
        };
    }
}
