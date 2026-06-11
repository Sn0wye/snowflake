namespace Oxygen.Domain;

public class ScoreTier
{
    public string Name { get; }
    public decimal Rate { get; }
    public decimal MaxLoanFraction { get; }

    private ScoreTier(string name, decimal rate, decimal maxLoanFraction)
    {
        Name = name;
        Rate = rate;
        MaxLoanFraction = maxLoanFraction;
    }

    public static ScoreTier For(int score)
    {
        if (score <= 599)
            return new ScoreTier("Poor", 0.22m, 0.15m);

        return score switch
        {
            <= 674 => new ScoreTier("Fair", 0.16m, 0.25m),
            <= 724 => new ScoreTier("Good", 0.12m, 0.35m),
            <= 774 => new ScoreTier("Very Good", 0.08m, 0.45m),
            <= 849 => new ScoreTier("Excellent", 0.05m, 0.55m),
            _ => new ScoreTier("Outstanding", 0.03m, 0.65m)
        };
    }
}
