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

        const int gameSessionId = 1;

        var townOfSalemSeeder = new TownOfSalemSeeder(rulesetId: 1, startingId: 1);
        townOfSalemSeeder
            .SeedData()
            .ApplyTo(builder);

        var blackJackSeeder = new BlackJackSeeder(rulesetId: 2, startingId: 100);
        blackJackSeeder
            .SeedData()
            .ApplyTo(builder);

        var playerSeeder = new PlayerSeeder();
        playerSeeder
            .SeedPlayers(gameSessionId)
            .AddRoles(townOfSalemSeeder.Roles)
            .ApplyTo(builder);        

        var gameSession = new GameSession
        {
            Id = gameSessionId,
            RulesetId = 1,
            Status = GameStatus.Running
        };

        var globalChat = new ChatChannel
        {
            Id = 1,
            Name = "Global",
            GameId = gameSessionId,
        };

        var globalChatMemberships = playerSeeder.Players.Select(p => new ChatChannelMembership
        {
            Player = p,
            PlayerId = p.Id,
            Channel = globalChat,
            ChannelId = globalChat.Id
        }).ToList();

        builder.Entity<GameSession>().HasData(gameSession);
        builder.Entity<ChatChannel>().HasData(globalChat);
    }
    public DbSet<Player> Players { get; set; }
    public DbSet<Role> Roles { get; set; }
    public DbSet<RoleKnowledge> RoleKnowledges { get; set; }
    public DbSet<Ability> Abilities { get; set; }
    public DbSet<Ruleset> Rulesets { get; set; }
    public DbSet<Round> Rounds { get; set; }
    public DbSet<GameSession> GameSessions { get; set; }
    public DbSet<ChatChannel> ChatChannels { get; set; }
    public DbSet<ChatMessage> ChatMessages { get; set; }

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
            .HasForeignKey<GameSession>(gs => gs.CurrentRoundId)
            .IsRequired(false);

        builder.Entity<GameSession>()
            .Property(gs => gs.Status)
            .HasConversion<int>();

        builder.Entity<Player>()
            .HasOne(p => p.GameSession)
            .WithMany(gs => gs.Participants)
            .HasForeignKey(p => p.GameId);

        builder.Entity<Player>()
            .HasMany(p => p.ChatChannelMemberships)
            .WithOne(ccm => ccm.Player)
            .IsRequired()
            .OnDelete(DeleteBehavior.Cascade);

        builder.Entity<ChatChannel>()
            .HasOne<GameSession>()
            .WithMany(gs => gs.ChatChannels)
            .HasForeignKey(cc => cc.GameId)
            .IsRequired()
            .OnDelete(DeleteBehavior.Cascade);

        builder.Entity<ChatChannelMembership>()
            .HasKey(ccm => new { ccm.PlayerId, ccm.ChannelId });

        builder.Entity<ChatChannelMembership>()
            .HasOne(ccm => ccm.Player)
            .WithMany(p => p.ChatChannelMemberships)
            .HasForeignKey(ccm => ccm.PlayerId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.Entity<ChatChannelMembership>()
             .HasOne(ccm => ccm.Channel)
             .WithMany(cc => cc.Members)
             .HasForeignKey(ccm => ccm.ChannelId)
             .OnDelete(DeleteBehavior.Cascade);

        builder.Entity<ChatMessage>()
            .HasOne(cm => cm.Channel)
            .WithMany(cc => cc.Messages)
            .HasForeignKey(cm => cm.ChannelId)
            .IsRequired()
            .OnDelete(DeleteBehavior.Cascade);

        builder.Entity<ChatMessage>()
            .HasOne(cm => cm.Sender)
            .WithMany()
            .HasForeignKey(cm => cm.SenderId)
            .IsRequired()
            .OnDelete(DeleteBehavior.Restrict);

        builder.Entity<ChatMessage>()
            .HasIndex(cm => new { cm.ChannelId, cm.SentAt })
            .HasDatabaseName("IX_ChatMessage_Channel_Time");

        ConfigureRoleAbilities(builder);
    }
}
