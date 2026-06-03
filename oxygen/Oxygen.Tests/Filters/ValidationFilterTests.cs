using System.Text.Json;
using FluentAssertions;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Abstractions;
using Microsoft.AspNetCore.Mvc.Filters;
using Microsoft.AspNetCore.Mvc.ModelBinding;
using Microsoft.AspNetCore.Routing;
using Oxygen.API.Filters;
using Xunit;

namespace Oxygen.Tests.Filters;

public class ValidationFilterTests
{
    private readonly ValidationFilter _sut = new();

    [Fact]
    public void does_nothing_when_model_state_is_valid()
    {
        var modelState = new ModelStateDictionary();
        var context = CreateActionExecutingContext(modelState);

        _sut.OnActionExecuting(context);

        context.Result.Should().BeNull();
    }

    [Fact]
    public void returns_400_with_errors_when_model_state_is_invalid()
    {
        var modelState = new ModelStateDictionary();
        modelState.AddModelError("LoanAmount", "Loan amount must be greater than 300.");
        modelState.AddModelError("Term", "Term is required.");
        var context = CreateActionExecutingContext(modelState);

        _sut.OnActionExecuting(context);

        var badRequest = context.Result.Should().BeOfType<BadRequestObjectResult>().Subject;
        var json = JsonSerializer.Serialize(badRequest.Value);
        var body = JsonSerializer.Deserialize<JsonElement>(json);

        body.GetProperty("message").GetString().Should().Be("Invalid request body.");
        body.GetProperty("status_code").GetInt32().Should().Be(400);

        var errors = body.GetProperty("errors");
        errors.GetProperty("LoanAmount").GetString().Should().NotBeNullOrWhiteSpace();
        errors.GetProperty("Term").GetString().Should().NotBeNullOrWhiteSpace();
    }

    [Fact]
    public void on_action_executed_is_noop()
    {
        var context = new ActionExecutedContext(
            new ActionContext(new DefaultHttpContext(), new RouteData(), new ActionDescriptor()),
            [],
            controller: null!);

        var act = () => _sut.OnActionExecuted(context);

        act.Should().NotThrow();
    }

    private static ActionExecutingContext CreateActionExecutingContext(ModelStateDictionary modelState)
    {
        return new ActionExecutingContext(
            new ActionContext(
                new DefaultHttpContext(),
                new RouteData(),
                new ActionDescriptor(),
                modelState),
            [],
            new Dictionary<string, object?>(),
            controller: null!);
    }
}
