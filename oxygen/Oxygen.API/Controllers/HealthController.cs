using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using Oxygen.Infrastructure;

namespace Oxygen.API.Controllers;

[ApiController]
[Route("loan")]
public class HealthController(ApplicationDbContext dbContext) : ControllerBase
{
    [HttpGet("health")]
    public async Task<IActionResult> Health()
    {
        var components = new Dictionary<string, string>();
        var allHealthy = true;

        // Database check
        try
        {
            var canConnect = await dbContext.Database.CanConnectAsync();
            if (canConnect)
            {
                components["database"] = "ok";
            }
            else
            {
                components["database"] = "cannot connect";
                allHealthy = false;
            }
        }
        catch (Exception e)
        {
            components["database"] = e.Message;
            allHealthy = false;
        }

        var response = new
        {
            status = allHealthy ? "healthy" : "unhealthy",
            components
        };

        return allHealthy
            ? Ok(response)
            : StatusCode(503, response);
    }
}
