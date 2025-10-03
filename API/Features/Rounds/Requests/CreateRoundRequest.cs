namespace API.Features.Rounds.Requests;
public readonly record struct CreateRoundRequest
{
    public DateTime StartTimeUTC { get; init; }
    public int DurationInSeconds { get; init; }
}
