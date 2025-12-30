namespace API.Features.Auth.Common;

public record struct AbilityInfo(int Id, string Name, string Description);
public record struct RoleInfo(int Id, string Name, string Description, List<AbilityInfo> Abilities);
