namespace Oxygen.Domain;

public class TermMultiplier
{
    public decimal Value { get; }

    private TermMultiplier(decimal value)
    {
        Value = value;
    }

    public static TermMultiplier For(int months)
    {
        var multiplier = months switch
        {
            <= 12 => 1.00m,
            <= 24 => 1.15m,
            <= 36 => 1.30m,
            <= 48 => 1.50m,
            _ => 1.70m
        };

        return new TermMultiplier(multiplier);
    }
}
