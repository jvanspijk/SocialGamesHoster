using API.Domain.Models;
using Microsoft.EntityFrameworkCore;
namespace API.DataAccess.Seeders;

internal class TownOfSalemSeeder(int rulesetId)
{
    private int _rulesetId = rulesetId;
    private Dictionary<string, Role> _roles = [];
    private Dictionary<string, Ability> _abilities = [];
    private List<object> _roleAbilities = [];
    private List<RoleKnowledge> _roleVisibilities = [];
    private bool _doneSeeding = false;
    public Ruleset? Ruleset { get; private set; }
    public List<Role> Roles => [.. _roles.Values];
    public List<Ability> Abilities => [.. _abilities.Values];

    public TownOfSalemSeeder SeedData()
    {
        SeedAbilities();
        SeedRoles();

        SeedRoleAbilities();
        SeedRoleVisibilities();

        SeedRuleset();        

        _doneSeeding = true;

        return this;
    }

    private void SeedRuleset()
    {
       Ruleset = new Ruleset()
       {
            Id = _rulesetId,
            Name = "Town of Salem",
            Description = "Town members aim to find and eliminate all evil roles, " +
                    "while Mafia and Neutral roles have their own secret goals, " +
                    "often involving the elimination of Town members. The game alternates between night, " +
                    "where players use their unique abilities, and day, where they discuss information, " +
                    "share their 'wills', and vote to hang someone.",
       };
    }

    private void SeedAbilities()
    {
        List<Ability> abilities = [
            new Ability { Id = 1, Name = "Vote", 
                Description = "Participate in daily voting to lynch a suspect."},
            new Ability { Id = 3, Name = "Defense", 
                Description = "Can defend themselves against night attacks."},
            new Ability { Id = 4, Name = "Heal", 
                Description = "Choose one player to protect from death each night."},
            new Ability { Id = 5, Name = "Investigate", 
                Description = "Choose one player to investigate each night and learn their role category." },
            new Ability { Id = 6, Name = "Shoot", 
                Description = "Choose one player to kill each night. Limited uses." },
            new Ability { Id = 7, Name = "Mafia Kill", 
                Description = "Execute the Mafia's chosen target at night." },
            new Ability { Id = 8, Name = "Organize Kill", 
                Description = "Choose the Mafia's nightly target. Appears as Townie to investigators." },
            new Ability { Id = 9, Name = "Trick", 
                Description = "If lynched, you will kill one player who voted for you that night." },
            new Ability { Id = 10, Name = "Target Elimination", 
                Description = "Win if your assigned target is lynched. You are immune to night kills until your target dies." },
        ];

        foreach (var ability in abilities)
        {
            ability.RulesetId = _rulesetId;
            _abilities[ability.Name] = ability;
        }
    }

    private void SeedRoles()
    {
        List<Role> roles = [
            new Role { Id = 1, Name = "Townie", Description = "A regular citizen of the town. Your goal is to eliminate all threats." },
            new Role { Id = 2, Name = "Doctor", Description = "You are a medical professional dedicated to saving lives." },
            new Role { Id = 3, Name = "Investigator", Description = "You seek the truth and uncover secrets hidden in the town." },
            new Role { Id = 4, Name = "Vigilante", Description = "You take justice into your own hands, even if it means getting your hands dirty." },
            new Role { Id = 5, Name = "Mafioso", Description = "A loyal member of the Mafia. You carry out the family's nightly kills." },
            new Role { Id = 6, Name = "Godfather", Description = "The cunning leader of the Mafia. You are immune to basic investigations." },
            new Role { Id = 7, Name = "Jester", Description = "Your only goal is to be lynched by the town." },
            new Role { Id = 8, Name = "Executioner", Description = "You have a specific target you must get lynched to win." },
        ];

        foreach (var role in roles)
        {
            role.RulesetId = _rulesetId;
            _roles[role.Name] = role;
        }
    }

    private void SeedRoleAbilities()
    {
        AddAbilityToRole("Townie", "Vote");

        AddAbilityToRole("Doctor", "Vote");
        AddAbilityToRole("Doctor", "Heal");

        AddAbilityToRole("Investigator", "Vote");
        AddAbilityToRole("Investigator", "Investigate");

        AddAbilityToRole("Vigilante", "Vote");
        AddAbilityToRole("Vigilante", "Shoot");

        AddAbilityToRole("Mafioso", "Vote");
        AddAbilityToRole("Mafioso", "Mafia Kill");

        AddAbilityToRole("Godfather", "Vote");
        AddAbilityToRole("Godfather", "Organize Kill");
        AddAbilityToRole("Godfather", "Defense");

        AddAbilityToRole("Jester", "Vote");
        AddAbilityToRole("Jester", "Trick");

        AddAbilityToRole("Executioner", "Vote");
        AddAbilityToRole("Executioner", "Target Elimination");
    }

    private void SeedRoleVisibilities()
    {
        AddRoleVisibility("Godfather", "Mafioso");
        AddRoleVisibility("Mafioso", "Godfather");
    }

    private void AddRoleVisibility(string seesName, string isSeenName)
    {
        if (!_roles.TryGetValue(seesName, out Role? role))
        {
            throw new InvalidOperationException($"Role '{seesName}' not found.");
        }
        if (!_roles.TryGetValue(isSeenName, out Role? isSeenRole))
        {
            throw new InvalidOperationException($"Role '{isSeenName}' not found.");
        }
        _roleVisibilities.Add(new RoleKnowledge{ SourceId = role.Id, TargetId = isSeenRole.Id, KnowledgeType = KnowledgeType.Role });
    }

    private void AddAbilityToRole(string roleName, string abilityName)
    {
        if (!_roles.TryGetValue(roleName, out Role? role))
        {
            throw new InvalidOperationException($"Role '{roleName}' not found.");
        }

        if (!_abilities.TryGetValue(abilityName, out Ability? ability))
        {
            throw new InvalidOperationException($"Ability '{abilityName}' not found.");
        }

        _roleAbilities.Add(new { RoleId = role.Id, AbilityId = ability.Id });
    }


    public void ApplyTo(ModelBuilder builder)
    {
        if (!_doneSeeding || Ruleset == null)
        {
            throw new InvalidOperationException("SeedData must be called before ApplyTo.");
        }

        builder.Entity<Ability>().HasData(Abilities);
        builder.Entity<Role>().HasData(Roles);

        builder.Entity<RoleKnowledge>().HasData(_roleVisibilities);
        builder.Entity<Role>().HasMany(r => r.Abilities).WithMany(a => a.AssociatedRoles)
            .UsingEntity("RoleAbilities",
                j =>
                {
                    j.HasKey("RoleId", "AbilityId");
                    j.HasData(_roleAbilities);
                });
            

        //builder.Entity("RoleAbility").HasData(_roleAbilities);

        builder.Entity<Ruleset>().HasData(Ruleset);
    }   
}
