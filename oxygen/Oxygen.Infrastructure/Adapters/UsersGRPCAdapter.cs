using Grpc.Net.Client;
using Microsoft.Extensions.Configuration;
using Pb;

namespace Oxygen.Infrastructure.Adapters;

public class UsersGRPCAdapter: IUsersGRPCAdapter
{
    private readonly UserService.UserServiceClient _client;

    public UsersGRPCAdapter(IConfiguration configuration)
    {
        var address = configuration.GetConnectionString("UsersGrpc");
        Console.WriteLine($"GRPC Address: {address}");
        if (string.IsNullOrWhiteSpace(address))
        {
            throw new InvalidOperationException("Configuration 'ConnectionStrings:UsersGrpc' is not set or is empty.");
        }
        var channel = GrpcChannel.ForAddress(address);
        _client = new UserService.UserServiceClient(channel);
    }
    
    public async Task<User> GetUserAsync(string userId)
    {
        var request = new GetUserRequest
        {
            Id = userId
        };
        var response =  await _client.GetUserAsync(request);
        return response.User;
    }
    
}

public interface IUsersGRPCAdapter
{
    Task<User> GetUserAsync(string userId);
}
