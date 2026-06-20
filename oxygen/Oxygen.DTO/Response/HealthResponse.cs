namespace Oxygen.DTO.Response;

public class HealthResponse
{
    /// <summary> Overall service status </summary>
    /// <example>healthy</example>
    public required string Status { get; set; }

    /// <summary> Per-dependency status </summary>
    public required HealthComponents Components { get; set; }
}

public class HealthComponents
{
    /// <summary> Database status ("ok" or error message) </summary>
    /// <example>ok</example>
    public required string Database { get; set; }
}
