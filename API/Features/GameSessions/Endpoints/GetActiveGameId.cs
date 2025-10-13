using API.DataAccess.Repositories;

namespace API.Features.GameSessions.Endpoints;
public static class GetActiveGameId
{
    public static IResult Handle()     
    {
        // TODO: inject repo and get actual active session id
        return Results.Ok(1); 
    }

}
