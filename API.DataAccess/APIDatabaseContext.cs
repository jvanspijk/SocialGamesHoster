using API.Domain.Models;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess;

public class APIDatabaseContext : DbContext
{
    public APIDatabaseContext(DbContextOptions options) : base(options) { }
    protected override void OnModelCreating(ModelBuilder builder)
    {
        base.OnModelCreating(builder);
        ConfigureEntityRelationships(builder);
        SeedGame(builder);
    }
    public DbSet<Player> Players { get; set; }
    public DbSet<Role> Roles { get; set; }
    public DbSet<Ability> Abilities { get; set; }
    public DbSet<Ruleset> Games { get; set; }

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

    private static void SeedGame(ModelBuilder builder)
    {
        const int gameId = 1;
        builder.Entity<Ruleset>().HasData(new Ruleset
        {
            Id = gameId,
            Name = "Town of Salem",
            Description = "Town members aim to find and eliminate all evil roles, " +
            "while Mafia and Neutral roles have their own secret goals, " +
            "often involving the elimination of Town members. The game alternates between night, " +
            "where players use their unique abilities, and day, where they discuss information, " +
            "share their 'wills', and vote to hang someone.",
        });
        List<Role> roles = SeedRoles(builder, gameId);
        SeedPlayers(builder, roles);        
    }

    private static void SeedPlayers(ModelBuilder builder, List<Role> roles)
    {
        int amountOfRoles = roles.Count;

        List<string> testUserNames = [
           "Alice", "John", "Emily", "Michael", "Sarah",
           "Jessica", "David", "Ashley", "Matthew", "Amanda",
           "Joshua", "Jennifer", "Daniel", "Elizabeth", "James",
           "Charlie", "Kyle", "Bob", "Megan", "Laura",
        ];

        int amountOfTestPlayers = testUserNames.Count;

        List<Player> players = new(amountOfTestPlayers);
        for (int i = 0; i < amountOfTestPlayers; i++)
        {
            int roleId = (i % amountOfRoles) + 1; // Cycle through roles

            string name = testUserNames[i] + " " + roles[i % amountOfRoles].Name;

            players.Add(new Player
            {
                Id = i + 1,
                Name = name,
                RoleId = roleId
            });
        }

        Player playerWithoutRole = new() { Id = amountOfTestPlayers + 1, Name = "User without role" };
        players.Add(playerWithoutRole);

        builder.Entity<Player>().HasData(players);
    }

    private static List<Role> SeedRoles(ModelBuilder builder, int gameId)
    {
        var basicVote = new Ability { Id = 1, Name = "Vote", Description = "Participate in daily voting to lynch a suspect.", GameId = gameId };
        var defense = new Ability { Id = 3, Name = "Defense", Description = "Can defend themselves against night attacks.", GameId = gameId };
        var heal = new Ability { Id = 4, Name = "Heal", Description = "Choose one player to protect from death each night.", GameId = gameId };
        var investigate = new Ability { Id = 5, Name = "Investigate", Description = "Choose one player to investigate each night and learn their role category.", GameId = gameId };
        var shoot = new Ability { Id = 6, Name = "Shoot", Description = "Choose one player to kill each night. Limited uses.", GameId = gameId };
        var mafiaKill = new Ability { Id = 7, Name = "Mafia Kill", Description = "Execute the Mafia's chosen target at night.", GameId = gameId };
        var organizeKill = new Ability { Id = 8, Name = "Organize Kill", Description = "Choose the Mafia's nightly target. Appears as Townie to investigators.", GameId = gameId };
        var trick = new Ability { Id = 9, Name = "Trick", Description = "If lynched, you will kill one player who voted for you that night.", GameId = gameId };
        var targetElimination = new Ability { Id = 10, Name = "Target Elimination", Description = "Win if your assigned target is lynched. You are immune to night kills until your target dies.", GameId = gameId };

        builder.Entity<Ability>().HasData(
            basicVote,
            defense,
            heal,
            investigate,
            shoot,
            mafiaKill,
            organizeKill,
            trick,
            targetElimination
        );

        var townie = new Role { Id = 1, Name = "Townie", Description = "A regular citizen of the town. Your goal is to eliminate all threats.", GameId = gameId };
        var doctor = new Role { Id = 2, Name = "Doctor", Description = "You are a medical professional dedicated to saving lives.", GameId = gameId };
        var investigator = new Role { Id = 3, Name = "Investigator", Description = "You seek the truth and uncover secrets hidden in the town.", GameId = gameId };
        var vigilante = new Role { Id = 4, Name = "Vigilante", Description = "You take justice into your own hands, even if it means getting your hands dirty.", GameId = gameId };
        var mafioso = new Role { Id = 5, Name = "Mafioso", Description = "A loyal member of the Mafia. You carry out the family's nightly kills.", GameId = gameId };
        var godfather = new Role { Id = 6, Name = "Godfather", Description = "The cunning leader of the Mafia. You are immune to basic investigations.", GameId = gameId };
        var jester = new Role { Id = 7, Name = "Jester", Description = "Your only goal is to be lynched by the town.", GameId = gameId };
        var executioner = new Role { Id = 8, Name = "Executioner", Description = "You have a specific target you must get lynched to win.", GameId = gameId };

        builder.Entity("RoleVisibility").HasData(
            new { RoleId = godfather.Id, VisibleRoleId = mafioso.Id },
            new { RoleId = mafioso.Id, VisibleRoleId = godfather.Id }
        );

        List<Role> roles = new()
        {
            townie,
            doctor,
            investigator,
            vigilante,
            mafioso,
            godfather,
            jester,
            executioner
        };

        builder.Entity<Role>().HasData(roles);

        builder.Entity("RoleAbility").HasData(
            // Townie abilities
            new { RoleId = townie.Id, AbilityId = basicVote.Id },

            // Doctor abilities
            new { RoleId = doctor.Id, AbilityId = basicVote.Id },
            new { RoleId = doctor.Id, AbilityId = heal.Id },

            // Investigator abilities
            new { RoleId = investigator.Id, AbilityId = basicVote.Id },
            new { RoleId = investigator.Id, AbilityId = investigate.Id },

            // Vigilante abilities
            new { RoleId = vigilante.Id, AbilityId = basicVote.Id },
            new { RoleId = vigilante.Id, AbilityId = shoot.Id },

            // Mafioso abilities
            new { RoleId = mafioso.Id, AbilityId = basicVote.Id },
            new { RoleId = mafioso.Id, AbilityId = mafiaKill.Id },

            // Godfather abilities
            new { RoleId = godfather.Id, AbilityId = basicVote.Id },
            new { RoleId = godfather.Id, AbilityId = organizeKill.Id },
            new { RoleId = godfather.Id, AbilityId = defense.Id },

            // Jester abilities
            new { RoleId = jester.Id, AbilityId = basicVote.Id },
            new { RoleId = jester.Id, AbilityId = trick.Id },

            // Executioner abilities
            new { RoleId = executioner.Id, AbilityId = basicVote.Id },
            new { RoleId = executioner.Id, AbilityId = targetElimination.Id }
        );

        return roles;
    }
}
