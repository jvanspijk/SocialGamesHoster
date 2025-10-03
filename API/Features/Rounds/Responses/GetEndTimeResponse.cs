namespace API.Features.Rounds.Responses;

public readonly record struct GetEndTimeResponse
{
    public DateTime EndTimeUTC { get; init; }
}
