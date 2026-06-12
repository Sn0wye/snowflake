using Grpc.Core;
using Grpc.Net.Client;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Configuration;
using Pb;

namespace Oxygen.Infrastructure.Adapters;

public class UsersGRPCAdapter: IUsersGRPCAdapter
{
    private readonly UserService.UserServiceClient _client;
    private readonly IHttpContextAccessor _httpContextAccessor;

    public UsersGRPCAdapter(IConfiguration configuration, IHttpContextAccessor httpContextAccessor)
    {
        var address = configuration.GetConnectionString("UsersGrpc");
        if (string.IsNullOrWhiteSpace(address))
        {
            throw new InvalidOperationException("Configuration 'ConnectionStrings:UsersGrpc' is not set or is empty.");
        }
        var channel = GrpcChannel.ForAddress(address);
        _client = new UserService.UserServiceClient(channel);
        _httpContextAccessor = httpContextAccessor;
    }
    
    public async Task<User> GetUserAsync(string userId)
    {
        var headers = new Metadata();

        var correlationId = _httpContextAccessor.HttpContext?.Items["correlation_id"] as string;
        if (!string.IsNullOrEmpty(correlationId))
        {
            headers.Add("x-correlation-id", correlationId);
        }

        var authHeader = _httpContextAccessor.HttpContext?.Request.Headers["Authorization"].FirstOrDefault();
        if (!string.IsNullOrEmpty(authHeader))
        {
            headers.Add("authorization", authHeader);
        }

        var request = new GetUserRequest
        {
            Id = userId
        };
        var response = await _client.GetUserAsync(request, headers);
        return response.User;
    }
    
}

public interface IUsersGRPCAdapter
{
    Task<User> GetUserAsync(string userId);
}
