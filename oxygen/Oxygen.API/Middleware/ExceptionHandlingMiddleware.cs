using System.Data.Common;
using Grpc.Core;
using Serilog.Context;

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
        var correlationId = context.Request.Headers["X-Correlation-ID"].FirstOrDefault();
        if (string.IsNullOrEmpty(correlationId))
        {
            correlationId = Guid.NewGuid().ToString();
        }
        context.Response.Headers["X-Correlation-ID"] = correlationId;
        context.Items["correlation_id"] = correlationId;

        using (LogContext.PushProperty("correlation_id", correlationId))
        {
            try
            {
                await _next(context);
            }
            catch (OperationCanceledException ex) when (ex.CancellationToken == context.RequestAborted)
            {
                _logger.LogInformation("Request cancelled by client");

                if (!context.Response.HasStarted)
                {
                    context.Response.StatusCode = 499;
                }
            }
            catch (OperationCanceledException)
            {
                throw;
            }
            catch (Exception ex) when (!context.Response.HasStarted)
            {
                _logger.LogError(ex, "Unhandled exception");
                await HandleExceptionAsync(context, ex);
            }
        }
    }

    private static async Task HandleExceptionAsync(HttpContext context, Exception exception)
    {
        var (statusCode, detail) = exception switch
        {
            RpcException rpc => MapGrpcStatus(rpc),
            DbException => (StatusCodes.Status503ServiceUnavailable, "Database operation failed."),
            HttpRequestException => (StatusCodes.Status502BadGateway, "External service call failed."),
            OperationCanceledException or TimeoutException => (StatusCodes.Status504GatewayTimeout, "Request timed out."),
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

    private static (int statusCode, string detail) MapGrpcStatus(RpcException rpc)
    {
        return rpc.StatusCode switch
        {
            StatusCode.NotFound => (StatusCodes.Status404NotFound, "Upstream resource not found."),
            StatusCode.InvalidArgument => (StatusCodes.Status400BadRequest, "Upstream gRPC invalid argument."),
            StatusCode.Unauthenticated => (StatusCodes.Status401Unauthorized, "Upstream gRPC authentication failed."),
            StatusCode.PermissionDenied => (StatusCodes.Status403Forbidden, "Upstream gRPC permission denied."),
            StatusCode.AlreadyExists => (StatusCodes.Status409Conflict, "Upstream resource already exists."),
            StatusCode.ResourceExhausted => (StatusCodes.Status429TooManyRequests, "Upstream resource exhausted."),
            StatusCode.Unavailable => (StatusCodes.Status503ServiceUnavailable, "Upstream gRPC service unavailable."),
            StatusCode.DeadlineExceeded => (StatusCodes.Status504GatewayTimeout, "Upstream gRPC call timed out."),
            _ => (StatusCodes.Status502BadGateway, "Upstream gRPC call failed.")
        };
    }
}
