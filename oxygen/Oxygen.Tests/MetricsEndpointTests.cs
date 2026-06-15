using System.Net;
using FluentAssertions;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Mvc.Testing;
using Xunit;

namespace Oxygen.Tests;

public class MetricsEndpointTests : IClassFixture<WebApplicationFactory<Program>>, IDisposable
{
    private readonly WebApplicationFactory<Program> _factory;

    public MetricsEndpointTests(WebApplicationFactory<Program> factory)
    {
        _factory = factory.WithWebHostBuilder(builder =>
        {
            builder.UseEnvironment("Production");
            // Real (non-default) secret so the JWT bootstrap check doesn't throw.
            builder.UseSetting("Security:Jwt:SecretKey", "test-secret-key-sufficiently-long-0123456789");
        });
    }

    [Fact]
    public async Task GetMetrics_ReturnsPrometheusTextWithHttpMetrics()
    {
        var client = _factory.CreateClient();

        // Generate traffic so the RED histogram has data.
        await client.GetAsync("/");

        var response = await client.GetAsync("/metrics");

        response.StatusCode.Should().Be(HttpStatusCode.OK);
        var body = await response.Content.ReadAsStringAsync();
        // RED: request rate/error-rate counter and latency histogram.
        body.Should().Contain("http_request_duration_seconds");
    }

    // WithWebHostBuilder returns a new factory instance that the IClassFixture
    // does not own, so dispose it here to tear down its host/TestServer.
    public void Dispose() => _factory.Dispose();
}
