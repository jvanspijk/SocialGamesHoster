namespace API.Features.Rulesets.Common;

public record struct AbilityInfo(int Id, string Name, string Description);
public record struct RoleInfo(int Id, string Name, string Description, List<int> AbilityIds, List<int> CanSeeRoleIds);
