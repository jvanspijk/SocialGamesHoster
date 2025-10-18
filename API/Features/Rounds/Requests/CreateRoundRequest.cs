namespace API.Features.Rounds.Requests;
public readonly record struct CreateRoundRequest
{
    public int DurationInSeconds { get; init; }
}
