using System.Data.Common;
using System.Text.Json;
using FluentAssertions;
using Grpc.Core;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Logging;
using Moq;
using Oxygen.API.Middleware;
using Xunit;

namespace Oxygen.Tests.Middleware;

public class ExceptionHandlingMiddlewareTests
{
    private readonly Mock<ILogger<ExceptionHandlingMiddleware>> _loggerMock;

    public ExceptionHandlingMiddlewareTests()
    {
        _loggerMock = new Mock<ILogger<ExceptionHandlingMiddleware>>();
    }

    [Fact]
    public async Task invokes_next_when_no_exception()
    {
        var context = new DefaultHttpContext();
        var called = false;
        RequestDelegate next = (ctx) => { called = true; return Task.CompletedTask; };
        var middleware = new ExceptionHandlingMiddleware(next, _loggerMock.Object);

        await middleware.InvokeAsync(context);

        called.Should().BeTrue();
        context.Response.StatusCode.Should().Be(200);
    }

    [Fact]
    public async Task returns_499_when_request_cancelled_by_client()
    {
        var context = new DefaultHttpContext();
        var cts = new CancellationTokenSource();
        cts.Cancel();
        context.RequestAborted = cts.Token;
        RequestDelegate next = (ctx) =>
            throw new OperationCanceledException("cancelled", cts.Token);
        var middleware = new ExceptionHandlingMiddleware(next, _loggerMock.Object);

        await middleware.InvokeAsync(context);

        context.Response.StatusCode.Should().Be(499);
    }

    [Theory]
    [InlineData(StatusCode.NotFound, 404)]
    [InlineData(StatusCode.InvalidArgument, 400)]
    [InlineData(StatusCode.Unauthenticated, 401)]
    [InlineData(StatusCode.PermissionDenied, 403)]
    [InlineData(StatusCode.AlreadyExists, 409)]
    [InlineData(StatusCode.ResourceExhausted, 429)]
    [InlineData(StatusCode.Unavailable, 503)]
    [InlineData(StatusCode.DeadlineExceeded, 504)]
    public async Task maps_grpc_status_codes_to_http_status_codes(
        StatusCode grpcCode, int expectedHttpCode)
    {
        var context = CreateResponseCaptureContext();
        var rpcEx = new RpcException(new Status(grpcCode, "grpc error"), new Metadata(), "detail");
        RequestDelegate next = (ctx) => throw rpcEx;
        var middleware = new ExceptionHandlingMiddleware(next, _loggerMock.Object);

        await middleware.InvokeAsync(context);

        context.Response.StatusCode.Should().Be(expectedHttpCode);
        var body = await ReadResponseBody(context);
        body.GetProperty("status_code").GetInt32().Should().Be(expectedHttpCode);
        body.GetProperty("message").GetString().Should().NotBeNullOrWhiteSpace();
    }

    [Fact]
    public async Task returns_503_when_database_exception()
    {
        var context = CreateResponseCaptureContext();
        RequestDelegate next = (ctx) =>
            throw new TestDbException("db error");
        var middleware = new ExceptionHandlingMiddleware(next, _loggerMock.Object);

        await middleware.InvokeAsync(context);

        context.Response.StatusCode.Should().Be(503);
        var body = await ReadResponseBody(context);
        body.GetProperty("status_code").GetInt32().Should().Be(503);
        body.GetProperty("message").GetString().Should().Be("Database operation failed.");
    }

    [Fact]
    public async Task returns_502_when_http_request_exception()
    {
        var context = CreateResponseCaptureContext();
        RequestDelegate next = (ctx) =>
            throw new HttpRequestException("external error");
        var middleware = new ExceptionHandlingMiddleware(next, _loggerMock.Object);

        await middleware.InvokeAsync(context);

        context.Response.StatusCode.Should().Be(502);
        var body = await ReadResponseBody(context);
        body.GetProperty("status_code").GetInt32().Should().Be(502);
        body.GetProperty("message").GetString().Should().Be("External service call failed.");
    }

    [Fact]
    public async Task returns_504_when_timeout_exception()
    {
        var context = CreateResponseCaptureContext();
        RequestDelegate next = (ctx) =>
            throw new TimeoutException("service timed out");
        var middleware = new ExceptionHandlingMiddleware(next, _loggerMock.Object);

        await middleware.InvokeAsync(context);

        context.Response.StatusCode.Should().Be(504);
    }

    [Fact]
    public async Task returns_500_when_unexpected_exception()
    {
        var context = CreateResponseCaptureContext();
        RequestDelegate next = (ctx) =>
            throw new InvalidOperationException("something broke");
        var middleware = new ExceptionHandlingMiddleware(next, _loggerMock.Object);

        await middleware.InvokeAsync(context);

        context.Response.StatusCode.Should().Be(500);
        var body = await ReadResponseBody(context);
        body.GetProperty("status_code").GetInt32().Should().Be(500);
        body.GetProperty("message").GetString().Should().Be("An unexpected error occurred.");
    }

    private static DefaultHttpContext CreateResponseCaptureContext()
    {
        var context = new DefaultHttpContext();
        context.Response.Body = new MemoryStream();
        return context;
    }

    private static async Task<JsonElement> ReadResponseBody(DefaultHttpContext context)
    {
        context.Response.Body.Position = 0;
        var text = await new StreamReader(context.Response.Body).ReadToEndAsync();
        return JsonSerializer.Deserialize<JsonElement>(text);
    }
}

public class TestDbException : DbException
{
    public TestDbException(string message) : base(message) { }
}
