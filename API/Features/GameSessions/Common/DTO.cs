namespace API.Features.GameSessions.Common;
public readonly record struct RoleInfo(int Id, string Name);
public readonly record struct Participant(int Id, string Name, RoleInfo? Role);
public readonly record struct Phase(int Id, string Name, string Description);