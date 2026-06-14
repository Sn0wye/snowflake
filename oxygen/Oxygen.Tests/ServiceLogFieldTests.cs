using FluentAssertions;
using Oxygen.API.Logging;
using Serilog;
using Serilog.Core;
using Serilog.Events;
using Xunit;

namespace Oxygen.Tests;

public class ServiceLogFieldTests
{
    private sealed class CapturingSink : ILogEventSink
    {
        public List<LogEvent> Events { get; } = new();
        public void Emit(LogEvent logEvent) => Events.Add(logEvent);
    }

    [Fact]
    public void ConfiguredLogger_EnrichesEveryEventWithServiceField()
    {
        var sink = new CapturingSink();
        var lc = new LoggerConfiguration();
        SerilogConfig.Configure(lc, isDevelopment: true);
        lc.WriteTo.Sink(sink);
        using var logger = lc.CreateLogger();

        logger.Information("probe");

        sink.Events.Should().ContainSingle();
        sink.Events[0].Properties.Should().ContainKey("service");
        sink.Events[0].Properties["service"].ToString().Should().Contain("oxygen");
    }
}
