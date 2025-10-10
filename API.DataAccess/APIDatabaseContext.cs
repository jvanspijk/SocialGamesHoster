using API.Domain.Models;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess;

public class APIDatabaseContext(DbContextOptions options) : DbContext(options)
{
    protected override void OnModelCreating(ModelBuilder builder)
    {
        base.OnModelCreating(builder);
        ConfigureEntityRelationships(builder);
        DataSeeder.SeedRuleset(builder);
    }
    public DbSet<Player> Players { get; set; }
    public DbSet<Role> Roles { get; set; }
    public DbSet<Ability> Abilities { get; set; }
    public DbSet<Ruleset> Rulesets { get; set; }
    public DbSet<Round> Rounds { get; set; }
    public DbSet<GameSession> GameSessions { get; set; }

    private static void ConfigureEntityRelationships(ModelBuilder builder)
    {
        builder.Entity<Role>()
            .HasMany(r => r.CanSee)
            .WithMany(r => r.CanBeSeenBy)
            .UsingEntity(r => r.ToTable("RoleVisibility"));                

        builder.Entity<Player>()
            .HasMany(p => p.CanSee)
            .WithMany(p => p.CanBeSeenBy)
            .UsingEntity(p => p.ToTable("PlayerVisibility"));

        builder.Entity<Role>()
            .HasMany(r => r.Abilities)
            .WithMany(a => a.AssociatedRoles)
            .UsingEntity(j => j.ToTable("RoleAbility"));

        builder.Entity<Player>()
            .HasOne(p => p.Role)
            .WithMany(r => r.PlayersWithRole)
            .HasForeignKey(p => p.RoleId)
            .IsRequired(false);
    }
}
