namespace API.Features.Players.Requests;
using API.Domain;
using API.Domain.Validation;

public readonly record struct DeletePlayerRequest : IValidatable<DeletePlayerRequest>
{
    public int Id { get; init; }
    public Result<DeletePlayerRequest> Validate(DeletePlayerRequest request)
    {
        if (request.Id <= 0)
        {
            return Errors.Validation("Id", "Player ID must be a positive integer.");
        }
        return request;
    }
}
