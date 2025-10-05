namespace API.Features.Rounds.Responses;

public readonly record struct RoundEndTimeResponse
{
    public DateTime EndTimeUTC { get; init; }
}
