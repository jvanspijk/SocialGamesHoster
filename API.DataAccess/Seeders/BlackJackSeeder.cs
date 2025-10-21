
using API.Domain.Models;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess.Seeders;

internal class BlackJackSeeder(int rulesetId, int startingId)
{
    private readonly int _idOffset = startingId;
    private readonly int _rulesetId = rulesetId;
    private readonly Dictionary<string, Role> _roles = [];
    private readonly Dictionary<string, Ability> _abilities = [];
    private readonly List<object> _roleAbilities = [];
    private readonly List<RoleKnowledge> _roleVisibilities = [];
    private bool _doneSeeding = false;
    public Ruleset? Ruleset { get; private set; }
    public List<Role> Roles => [.. _roles.Values];
    public List<Ability> Abilities => [.. _abilities.Values];
    public BlackJackSeeder SeedData()
    {
        SeedAbilities();
        SeedRoles();

        SeedRoleAbilities();
        SeedRoleVisibilities();

        SeedRuleset();

        _doneSeeding = true;

        return this;
    }

    public void ApplyTo(ModelBuilder builder)
    {
        if (!_doneSeeding)
        {
            throw new InvalidOperationException("Seeder has not been seeded yet. Call SeedData() before applying to ModelBuilder.");
        }
        builder.Entity<Ruleset>().HasData(Ruleset!);
        builder.Entity<Role>().HasData(Roles);
        builder.Entity<Ability>().HasData(Abilities);
        builder.Entity<Role>()
            .HasMany(r => r.Abilities)
            .WithMany(a => a.AssociatedRoles)
            .UsingEntity("RoleAbilities",
                j =>
                {
                    j.HasKey("RoleId", "AbilityId");
                    j.HasData(_roleAbilities);
                });
        builder.Entity<RoleKnowledge>().HasData(_roleVisibilities);
    }

    private void SeedRoleVisibilities()
    {
        // In Blackjack, everyone knows each other's roles
        foreach (var roleA in _roles.Values)
        {
            foreach (var roleB in _roles.Values)
            {
                _roleVisibilities.Add(new RoleKnowledge
                {
                    SourceId = roleA.Id,
                    TargetId = roleB.Id,
                    KnowledgeType = KnowledgeType.Role
                });
            }
        }
    }

    private void SeedRuleset()
    {
        Ruleset = new Ruleset()
        {
            Id = _rulesetId,
            Name = "Blackjack",
            Description = "Blackjack, also known as 21, is a popular card game where players aim to have a hand value as close to 21 as possible without exceeding it. " +
            "Players compete against the dealer rather than each other. " +
            "The game involves strategic decisions such as hitting, standing, doubling down, and splitting pairs.",
        };
    }

    private void SeedRoleAbilities()
    {
        // Player Abilities
        _roleAbilities.Add(new { RoleId = _roles["Player"].Id, AbilityId = _abilities["Hit"].Id });
        _roleAbilities.Add(new { RoleId = _roles["Player"].Id, AbilityId = _abilities["Stand"].Id });
        _roleAbilities.Add(new { RoleId = _roles["Player"].Id, AbilityId = _abilities["Double Down"].Id });
        _roleAbilities.Add(new { RoleId = _roles["Player"].Id, AbilityId = _abilities["Split"].Id });

        // Dealer Abilities
        _roleAbilities.Add(new { RoleId = _roles["Dealer"].Id, AbilityId = _abilities["Hit"].Id });
        _roleAbilities.Add(new { RoleId = _roles["Dealer"].Id, AbilityId = _abilities["Stand"].Id });
    }

    private void SeedRoles()
    {
        List<Role> roles = [
            new Role { Id = 1, Name = "Player", 
                Description = "A participant in the game aiming to beat the dealer by achieving a hand value of 21 or less."},
            new Role { Id = 2, Name = "Dealer", 
                Description = "The house representative who manages the game, deals cards, and plays against the players."},
        ];

        foreach (var role in roles)
        {
            role.Id += _idOffset;
            role.RulesetId = _rulesetId;
            _roles[role.Name] = role;
        }
    }

    private void SeedAbilities()
    {
        List<Ability> abilities = [
            new Ability { Id = 1, Name = "Hit", 
                Description = "Request an additional card to increase hand value."},
            new Ability { Id = 2, Name = "Stand", 
                Description = "Keep current hand and end turn."},
            new Ability { Id = 3, Name = "Double Down", 
                Description = "Double the initial bet and receive one final card."},
            new Ability { Id = 4, Name = "Split", 
                Description = "If dealt a pair, split into two separate hands." },
        ];

        foreach (var ability in abilities)
        {
            ability.Id += _idOffset;
            ability.RulesetId = _rulesetId;
            _abilities[ability.Name] = ability;
        }
    }
}
