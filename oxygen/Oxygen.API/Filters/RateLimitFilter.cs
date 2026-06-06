using System.Collections.Concurrent;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Filters;

namespace Oxygen.API.Filters;

public class RateLimitFilter : IActionFilter
{
    private static readonly ConcurrentDictionary<string, SlidingWindow> Buckets = new();
    private static readonly Timer ReapTimer;

    private readonly int _rate;

    static RateLimitFilter()
    {
        ReapTimer = new Timer(_ => Reap(), null, TimeSpan.FromMinutes(1), TimeSpan.FromMinutes(1));
    }

    public RateLimitFilter(IConfiguration configuration)
    {
        _rate = configuration.GetValue<int>("RateLimiting:Loan:PermitLimit", 10);
        if (_rate <= 0) _rate = 10;
    }

    public void OnActionExecuting(ActionExecutingContext context)
    {
        var enabled = context.HttpContext.RequestServices
            .GetRequiredService<IConfiguration>()
            .GetValue<bool>("RateLimiting:Enabled", true);

        if (!enabled) return;

        var userId = context.HttpContext.User.FindFirst(
            System.Security.Claims.ClaimTypes.NameIdentifier)?.Value;

        if (userId == null) return;

        var window = Buckets.GetOrAdd(userId, _ => new SlidingWindow());
        window.Touch();
        if (!window.Allow(_rate))
        {
            context.HttpContext.Response.Headers["Retry-After"] = "1";
            context.Result = new ObjectResult(new
            {
                success = false,
                message = "Too many requests. Try again in 1 second.",
                retry_after = 1
            })
            {
                StatusCode = 429
            };
        }
    }

    public void OnActionExecuted(ActionExecutedContext context) { }

    private static void Reap()
    {
        var cutoff = DateTimeOffset.UtcNow.AddMinutes(-1).ToUnixTimeMilliseconds();
        foreach (var key in Buckets.Keys)
        {
            if (Buckets.TryGetValue(key, out var window) && window.LastSeen < cutoff)
            {
                Buckets.TryRemove(key, out _);
            }
        }
    }

    public void Dispose()
    {
        ReapTimer?.Dispose();
    }

    private class SlidingWindow
    {
        private long _windowStart = DateTimeOffset.UtcNow.Ticks;
        private int _count;
        public long LastSeen { get; private set; }

        public SlidingWindow()
        {
            Touch();
        }

        public void Touch()
        {
            LastSeen = DateTimeOffset.UtcNow.ToUnixTimeMilliseconds();
        }

        public bool Allow(int rate)
        {
            lock (this)
            {
                var now = DateTimeOffset.UtcNow.Ticks;
                if (now - _windowStart > TimeSpan.TicksPerSecond)
                {
                    _windowStart = now;
                    _count = 0;
                }
                if (_count >= rate) return false;
                _count++;
                return true;
            }
        }
    }
}
