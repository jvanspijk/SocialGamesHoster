using API.Domain;
using API.Domain.Validation;

namespace API.Features.Players.Requests;
public record class CreatePlayerRequest(string Name) : IValidatable<CreatePlayerRequest>
{
    public Result<CreatePlayerRequest> Validate(CreatePlayerRequest request)
    {
        if (string.IsNullOrWhiteSpace(request.Name))
        {
            return Errors.Validation("Name", "Player name must not be empty.");
        }
        if (request.Name.Length < 3 || request.Name.Length > 20)
        {
            return Errors.Validation("Name", "Player name must be between 3 and 20 characters long.");
        }
        return request;
    }
}

