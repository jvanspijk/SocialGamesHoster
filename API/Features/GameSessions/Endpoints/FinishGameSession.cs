﻿using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;

namespace API.Features.GameSessions.Endpoints;

public static class FinishGameSession
{
    public static async Task<IResult> HandleAsync(GameSessionRepository repository, int gameId)
    {
        Result<GameSession> result = await repository.FinishGameSession(gameId);
        if(!result.IsSuccess)
        {
            return result.AsIResult();
        }

        return Results.NoContent();
    }
}