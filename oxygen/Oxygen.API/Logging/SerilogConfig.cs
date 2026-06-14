using Serilog;
using Serilog.Formatting.Compact;

namespace Oxygen.API.Logging;

/// <summary>
/// Central Serilog setup so the logging configuration can be exercised in tests.
/// </summary>
public static class SerilogConfig
{
    public static void Configure(LoggerConfiguration lc, bool isDevelopment)
    {
        lc.Enrich.FromLogContext()
            // Identify the origin service on every log line for log aggregation.
            .Enrich.WithProperty("service", "oxygen");

        if (isDevelopment)
            lc.WriteTo.Console();
        else
            lc.WriteTo.Console(new CompactJsonFormatter());
    }
}
