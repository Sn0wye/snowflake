using Swashbuckle.AspNetCore.Annotations;

namespace Oxygen.DTO.Response;

public class ValidationErrorResponse
{
    // <summary>
    // Error message
    // </summary>
    // <example>Invalid request body.</example>
    public required string Message { get; set; }
    
    // <summary>
    // Status code
    // </summary>
    // <example>400</example>
    public required int StatusCode { get; set; }
    
    // <summary>
    // Validation errors
    // </summary>
    // <example>{"loanAmount": "The LoanAmount field is required.", "term": "Term must be between 1 and 12." }</example>
    public required Dictionary<string, string> Errors { get; set; }
}