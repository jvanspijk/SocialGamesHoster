using API.Domain;
using Microsoft.AspNetCore.Http.HttpResults;

namespace API;

internal static class APIResults
{
    // 200
    internal static Ok Ok()
        => TypedResults.Ok();

    internal static Ok<T> Ok<T>(T value)
        => TypedResults.Ok(value);

    // 201
    internal static CreatedAtRoute<T> CreatedAtRoute<T>(T value, string getRouteName, int id)
        => TypedResults.CreatedAtRoute(value, getRouteName, new { id });

    // 204
    internal static NoContent NoContent()
        => TypedResults.NoContent();

    // 400
    internal static ProblemHttpResult BadRequest(string detail)
        => TypedResults.Problem(detail: detail, statusCode: 400);

    // 401
    internal static ProblemHttpResult Unauthorized(string detail = "Authentication is required to access this resource.")
        => TypedResults.Problem(detail: detail, statusCode: StatusCodes.Status401Unauthorized);

    // 403 - Permission Denied
    internal static ProblemHttpResult Forbidden(string detail = "You do not have the necessary permissions to perform this action.")
        => TypedResults.Problem(detail: detail, statusCode: StatusCodes.Status403Forbidden);

    // 404
    internal static ProblemHttpResult NotFound(string detail)
        => TypedResults.Problem(detail: detail, statusCode: 404);

    internal static ProblemHttpResult NotFound<T>(int id)
        => TypedResults.Problem(
            detail: $"Could not find an instance of `{nameof(T)}` with id `{id}`",
            statusCode: 404);

    // 409
    internal static ProblemHttpResult Conflict(string detail)
        => TypedResults.Problem(detail: detail, statusCode: 409);

    // 500
    internal static ProblemHttpResult InternalServerError(string detail)
        => TypedResults.Problem(detail: detail, statusCode: 500);
}
