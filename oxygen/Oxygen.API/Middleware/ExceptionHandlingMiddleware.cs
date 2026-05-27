using System.Data.Common;

namespace Oxygen.API.Middleware;

public class ExceptionHandlingMiddleware
{
    private readonly RequestDelegate _next;
    private readonly ILogger<ExceptionHandlingMiddleware> _logger;

    public ExceptionHandlingMiddleware(RequestDelegate next, ILogger<ExceptionHandlingMiddleware> logger)
    {
        _next = next;
        _logger = logger;
    }

    public async Task InvokeAsync(HttpContext context)
    {
        try
        {
            await _next(context);
        }
        catch (Exception ex) when (!context.Response.HasStarted)
        {
            _logger.LogError(ex, "Unhandled exception");
            await HandleExceptionAsync(context, ex);
        }
    }

    private static async Task HandleExceptionAsync(HttpContext context, Exception exception)
    {
        var (statusCode, detail) = exception switch
        {
            DbException => (StatusCodes.Status503ServiceUnavailable, "Database operation failed."),
            HttpRequestException => (StatusCodes.Status502BadGateway, "External service call failed."),
            TimeoutException or TaskCanceledException => (StatusCodes.Status504GatewayTimeout, "Request timed out."),
            ArgumentException => (StatusCodes.Status400BadRequest, exception.Message),
            _ when IsGrpcException(exception) => (StatusCodes.Status502BadGateway, "Upstream gRPC call failed."),
            _ => (StatusCodes.Status500InternalServerError, "An unexpected error occurred.")
        };

        context.Response.ContentType = "application/json";
        context.Response.StatusCode = statusCode;

        var response = new
        {
            message = detail,
            status_code = statusCode
        };

        await context.Response.WriteAsJsonAsync(response);
    }

    private static bool IsGrpcException(Exception exception)
    {
        return exception.GetType().FullName == "Grpc.Core.RpcException";
    }
}
