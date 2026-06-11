namespace Oxygen.Domain;

public static class DecimalPower
{
    public static decimal Pow(decimal baseValue, int exponent)
    {
        if (exponent < 0)
            throw new ArgumentOutOfRangeException(nameof(exponent),
                $"Exponent must be non-negative, got {exponent}.");

        var result = 1m;
        var b = baseValue;
        var e = exponent;

        while (e > 0)
        {
            if ((e & 1) == 1)
                result *= b;
            b *= b;
            e >>= 1;
        }

        return result;
    }
}
