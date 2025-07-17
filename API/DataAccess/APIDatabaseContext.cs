using API.Models;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess;

public class APIDatabaseContext : DbContext
{
    public APIDatabaseContext(DbContextOptions options) : base(options) { }

    protected override void OnModelCreating(ModelBuilder builder)
    {
        base.OnModelCreating(builder);
        ConfigureEntityRelationships(builder);
        SeedData(builder);
    }
    public DbSet<Player> Players { get; set; }
    public DbSet<Role> Roles { get; set; }
    public DbSet<Ability> Abilities { get; set; }
    public DbSet<RoleAbility> RoleAbilityAssociations { get; set; }

    private void ConfigureEntityRelationships(ModelBuilder builder)
    {
        builder.Entity<RoleAbility>()
            .HasKey(ra => new { ra.RoleId, ra.AbilityId });

        builder.Entity<RoleAbility>()
            .HasOne(ra => ra.Role)
            .WithMany(r => r.AbilityAssociations)
            .HasForeignKey(ra => ra.RoleId);

        builder.Entity<RoleAbility>()
            .HasOne(ra => ra.Ability)
            .WithMany(a => a.RoleAssociations)
            .HasForeignKey(ra => ra.AbilityId);

        builder.Entity<Player>()
            .HasOne(p => p.Role)
            .WithMany()
            .HasForeignKey(p => p.RoleId)
            .IsRequired(false);
    }

    private void SeedData(ModelBuilder builder)
    {
        SeedPlayers(builder);
        SeedRolesAndAbilities(builder);
    }

    private void SeedPlayers(ModelBuilder builder)
    {
        Random random = new();
        int amountOfRoles = 8; // Ugly hardcoded, but this is just for testing purposes
        
        List<string> testUserNames = [
           "Alice", "John", "Emily", "Michael", "Sarah",
           "Jessica", "David", "Ashley", "Matthew", "Amanda",
           "Joshua", "Jennifer", "Daniel", "Elizabeth", "James",
           "Charlie", "Kyle", "Bob", "Megan", "Laura",
        ];

        List<Player> players = new(testUserNames.Count);
        for (int i = 0; i < testUserNames.Count; i++)
        {
            int roleId = random.Next(1, amountOfRoles + 1);

            players.Add(new Player
            {
                Id = i + 1,
                Name = testUserNames[i],
                RoleId = roleId
            });
        }

        builder.Entity<Player>().HasData(players);
    }

    private void SeedRolesAndAbilities(ModelBuilder builder)
    {
        var basicVote = new Ability { Id = 1, Name = "Vote", Description = "Participate in daily voting to lynch a suspect." };
        var defense = new Ability { Id = 3, Name = "Defense", Description = "Can defend themselves against night attacks." };
        var heal = new Ability { Id = 4, Name = "Heal", Description = "Choose one player to protect from death each night." };
        var investigate = new Ability { Id = 5, Name = "Investigate", Description = "Choose one player to investigate each night and learn their role category." };
        var shoot = new Ability { Id = 6, Name = "Shoot", Description = "Choose one player to kill each night. Limited uses." };
        var mafiaKill = new Ability { Id = 7, Name = "Mafia Kill", Description = "Execute the Mafia's chosen target at night." };
        var organizeKill = new Ability { Id = 8, Name = "Organize Kill", Description = "Choose the Mafia's nightly target. Appears as Townie to investigators." };
        var trick = new Ability { Id = 9, Name = "Trick", Description = "If lynched, you will kill one player who voted for you that night." };
        var targetElimination = new Ability { Id = 10, Name = "Target Elimination", Description = "Win if your assigned target is lynched. You are immune to night kills until your target dies." };

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

        var townie = new Role { Id = 1, Name = "Townie", Description = "A regular citizen of the town. Your goal is to eliminate all threats." };
        var doctor = new Role { Id = 2, Name = "Doctor", Description = "You are a medical professional dedicated to saving lives." };
        var investigator = new Role { Id = 3, Name = "Investigator", Description = "You seek the truth and uncover secrets hidden in the town." };
        var vigilante = new Role { Id = 4, Name = "Vigilante", Description = "You take justice into your own hands, even if it means getting your hands dirty." };
        var mafioso = new Role { Id = 5, Name = "Mafioso", Description = "A loyal member of the Mafia. You carry out the family's nightly kills." };
        var godfather = new Role { Id = 6, Name = "Godfather", Description = "The cunning leader of the Mafia. You are immune to basic investigations." };
        var jester = new Role { Id = 7, Name = "Jester", Description = "Your only goal is to be lynched by the town." };
        var executioner = new Role { Id = 8, Name = "Executioner", Description = "You have a specific target you must get lynched to win." };

        builder.Entity<Role>().HasData(
            townie,
            doctor,
            investigator,
            vigilante,
            mafioso,
            godfather,
            jester,
            executioner
        );

        builder.Entity<RoleAbility>().HasData(
            // Townie abilities
            new RoleAbility { RoleId = townie.Id, AbilityId = basicVote.Id },

            // Doctor abilities
            new RoleAbility { RoleId = doctor.Id, AbilityId = basicVote.Id },
            new RoleAbility { RoleId = doctor.Id, AbilityId = heal.Id },

            // Investigator abilities
            new RoleAbility { RoleId = investigator.Id, AbilityId = basicVote.Id },
            new RoleAbility { RoleId = investigator.Id, AbilityId = investigate.Id },

            // Vigilante abilities
            new RoleAbility { RoleId = vigilante.Id, AbilityId = basicVote.Id },
            new RoleAbility { RoleId = vigilante.Id, AbilityId = shoot.Id },

            // Mafioso abilities
            new RoleAbility { RoleId = mafioso.Id, AbilityId = basicVote.Id },
            new RoleAbility { RoleId = mafioso.Id, AbilityId = mafiaKill.Id },

            // Godfather abilities
            new RoleAbility { RoleId = godfather.Id, AbilityId = basicVote.Id },
            new RoleAbility { RoleId = godfather.Id, AbilityId = organizeKill.Id },
            new RoleAbility { RoleId = godfather.Id, AbilityId = defense.Id },

            // Jester abilities
            new RoleAbility { RoleId = jester.Id, AbilityId = basicVote.Id },
            new RoleAbility { RoleId = jester.Id, AbilityId = trick.Id },

            // Executioner abilities
            new RoleAbility { RoleId = executioner.Id, AbilityId = basicVote.Id },
            new RoleAbility { RoleId = executioner.Id, AbilityId = targetElimination.Id }
        );
    }
}
