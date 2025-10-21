namespace API.Features.GameSessions.Common;
public record struct RoleInfo(int Id, string Name);
public record struct Participant(int Id, string Name, RoleInfo? Role);