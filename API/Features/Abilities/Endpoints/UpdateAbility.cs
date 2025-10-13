namespace API.Features.Abilities.Endpoints;

public static class UpdateAbility
{
    public readonly record struct Request(int Id, string? NewName, string? NewDescription);

    public static async Task<IResult> HandleAsync(AbilityRepository repository, Request request)
    {
        Ability? ability = await repository.GetByIdAsync(request.Id);
        if (ability == null)
        {
            return Results.NotFound();
        }
        if (!string.IsNullOrWhiteSpace(request.NewName))
        {
            ability.Name = request.NewName;
        }
        if (!string.IsNullOrWhiteSpace(request.NewDescription))
        {
            ability.Description = request.NewDescription;
        }
        Ability updatedAbility = await repository.UpdateAsync(ability);
        AbilityResponse response = new(updatedAbility.Id, updatedAbility.Name, updatedAbility.Description);
        return Results.Ok(response);
    }
}
