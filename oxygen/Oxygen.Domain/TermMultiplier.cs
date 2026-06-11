namespace Oxygen.Domain;

public class TermMultiplier
{
    public double Value { get; }

    private TermMultiplier(double value)
    {
        Value = value;
    }

    public static TermMultiplier For(int months)
    {
        var multiplier = months switch
        {
            <= 12 => 1.00,
            <= 24 => 1.15,
            <= 36 => 1.30,
            <= 48 => 1.50,
            _ => 1.70
        };

        return new TermMultiplier(multiplier);
    }
}
