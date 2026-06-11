namespace Oxygen.Domain;

public class ScoreTier
{
    public string Name { get; }
    public double BaseRate { get; }
    public double MaxLoanPercentage { get; }

    private ScoreTier(string name, double baseRate, double maxLoanPercentage)
    {
        Name = name;
        BaseRate = baseRate;
        MaxLoanPercentage = maxLoanPercentage;
    }

    public static ScoreTier For(int score)
    {
        if (score <= 599)
            return new ScoreTier("Poor", 22, 0.15);

        return score switch
        {
            <= 674 => new ScoreTier("Fair", 16, 0.25),
            <= 724 => new ScoreTier("Good", 12, 0.35),
            <= 774 => new ScoreTier("Very Good", 8, 0.45),
            <= 849 => new ScoreTier("Excellent", 5, 0.55),
            _ => new ScoreTier("Outstanding", 3, 0.65)
        };
    }
}
