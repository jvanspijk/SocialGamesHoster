using API.DataAccess.Seeders;
using API.Domain.Models;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess;

public class APIDatabaseContext(DbContextOptions options) : DbContext(options)
{
    protected override void OnModelCreating(ModelBuilder builder)
    {
        base.OnModelCreating(builder);
        ConfigureEntityRelationships(builder);

        const int rulesetId = 1;
        const int gameSessionId = 1;

        var rulesetSeeder = new TownOfSalemSeeder(rulesetId);
        rulesetSeeder
            .SeedData()
            .ApplyTo(builder);

        var playerSeeder = new PlayerSeeder();
        playerSeeder
            .SeedPlayers(gameSessionId)
            .AddRoles(rulesetSeeder.Roles)
            .ApplyTo(builder);        

        var gameSession = new GameSession
        {
            Id = gameSessionId,
            RulesetId = rulesetId,
            Status = GameStatus.Running
        };

        builder.Entity<GameSession>().HasData(gameSession);
    }
    public DbSet<Player> Players { get; set; }
    public DbSet<Role> Roles { get; set; }
    public DbSet<RoleKnowledge> RoleKnowledges { get; set; }
    public DbSet<Ability> Abilities { get; set; }
    public DbSet<Ruleset> Rulesets { get; set; }
    public DbSet<Round> Rounds { get; set; }
    public DbSet<GameSession> GameSessions { get; set; }

    private static void ConfigureRoleAbilities(ModelBuilder builder)
    {
        builder.Entity<Role>()
        .HasMany(r => r.Abilities)
        .WithMany(a => a.AssociatedRoles)
        .UsingEntity(
            "RoleAbilities",
            j =>
            {
                j.Property<int>("RoleId");
                j.Property<int>("AbilityId");            
            });
        // The rest of the configuration happens in the seeder,
        // because implicit join entities have to be seeded while defining the relationship
    }

    private static void ConfigureEntityRelationships(ModelBuilder builder)
    {
        builder.Entity<RoleKnowledge>()
            .HasKey(rv => new { rv.SourceId, rv.TargetId });

        builder.Entity<RoleKnowledge>()
            .HasOne(rv => rv.Source)
            .WithMany(r => r.KnowsAbout)
            .HasForeignKey(rv => rv.SourceId)
            .IsRequired();

        builder.Entity<RoleKnowledge>()
            .HasOne(rv => rv.Target)
            .WithMany(r => r.KnownBy)
            .HasForeignKey(rv => rv.TargetId)
            .IsRequired();

        builder.Entity<Player>()
            .HasMany(p => p.CanSee)
            .WithMany(p => p.CanBeSeenBy)
            .UsingEntity(p => p.ToTable("PlayerVisibility"));

        builder.Entity<Player>()
            .HasOne(p => p.Role)
            .WithMany(r => r.PlayersWithRole)
            .HasForeignKey(p => p.RoleId)
            .IsRequired(false);

        builder.Entity<Round>()
            .HasOne(r => r.GameSession)              
            .WithMany(gs => gs.Rounds)
            .HasForeignKey(r => r.GameId)
            .IsRequired()
            .OnDelete(DeleteBehavior.Cascade);

        builder.Entity<GameSession>()
            .HasOne(gs => gs.CurrentRound)
            .WithOne()
            .HasForeignKey<GameSession>()
            .IsRequired(false);

        builder.Entity<GameSession>()
            .Property(gs => gs.Status)
            .HasConversion<int>();

        builder.Entity<Player>()
            .HasOne(p => p.GameSession)
            .WithMany(gs => gs.Participants)
            .HasForeignKey(p => p.GameId);

        ConfigureRoleAbilities(builder);
    }
}
