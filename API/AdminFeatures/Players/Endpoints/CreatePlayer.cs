namespace API.AdminFeatures.Players.Endpoints;
public static class CreatePlayer
{
    public static async Task<IResult> HandleAsync(AdminPlayerService playerService)
    {
        var result = await playerService.GetAllAsync();
        return result.AsIResult();
    }
}
