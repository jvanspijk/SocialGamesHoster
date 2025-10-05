namespace API.AdminFeatures.Players.Endpoints;

public static class GetPlayers
{
    public static async Task<IResult> HandleAsync(AdminPlayerService playerService, int id)
    {
        var result = await playerService.GetAllAsync();
        return result.AsIResult();
    }
}