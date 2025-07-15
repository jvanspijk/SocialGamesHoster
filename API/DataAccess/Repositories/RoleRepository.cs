using API.Models;
using API.Validation;
using LanguageExt.Common;
using System.Linq;

namespace API.DataAccess.Repositories;

public class RoleRepository
{
    private List<Role> _roles;

    public RoleRepository() {
        _roles = new List<Role>();

        // Define common abilities that can be shared
        Ability basicVote = new Ability { Name = "Vote", Description = "Participate in daily voting to lynch a suspect." };
        Ability nightAction = new Ability { Name = "Night Action", Description = "Perform a specific action during the night phase." };
        Ability defense = new Ability { Name = "Defense", Description = "Can defend themselves against night attacks." };

        // Create roles and assign descriptions and specific abilities
        Role townie = new Role
        {
            Id = 1,
            Name = "Townie",
            Description = "A regular citizen of the town. Your goal is to eliminate all threats."
        };
        townie.Abilities.Add(basicVote);

        Role doctor = new Role
        {
            Id = 2,
            Name = "Doctor",
            Description = "You are a medical professional dedicated to saving lives."
        };
        doctor.Abilities.Add(basicVote);
        doctor.Abilities.Add(new Ability { Name = "Heal", Description = "Choose one player to protect from death each night." });

        Role investigator = new Role
        {
            Id = 3,
            Name = "Investigator",
            Description = "You seek the truth and uncover secrets hidden in the town."
        };
        investigator.Abilities.Add(basicVote);
        investigator.Abilities.Add(new Ability { Name = "Investigate", Description = "Choose one player to investigate each night and learn their role category." });

        Role vigilante = new Role
        {
            Id = 4,
            Name = "Vigilante",
            Description = "You take justice into your own hands, even if it means getting your hands dirty."
        };
        vigilante.Abilities.Add(basicVote);
        vigilante.Abilities.Add(new Ability { Name = "Shoot", Description = "Choose one player to kill each night. Limited uses." });

        Role mafioso = new Role
        {
            Id = 5,
            Name = "Mafioso",
            Description = "A loyal member of the Mafia. You carry out the family's nightly kills."
        };
        mafioso.Abilities.Add(basicVote);
        mafioso.Abilities.Add(new Ability { Name = "Mafia Kill", Description = "Execute the Mafia's chosen target at night." });

        Role godfather = new Role
        {
            Id = 6,
            Name = "Godfather",
            Description = "The cunning leader of the Mafia. You are immune to basic investigations."
        };
        godfather.Abilities.Add(basicVote);
        godfather.Abilities.Add(new Ability { Name = "Organize Kill", Description = "Choose the Mafia's nightly target. Appears as Townie to investigators." });
        godfather.Abilities.Add(defense); // Example: Godfather might have basic defense

        Role jester = new Role
        {
            Id = 7,
            Name = "Jester",
            Description = "Your only goal is to be lynched by the town."
        };
        jester.Abilities.Add(basicVote);
        jester.Abilities.Add(new Ability { Name = "Trick", Description = "If lynched, you will kill one player who voted for you that night." });

        Role executioner = new Role
        {
            Id = 8,
            Name = "Executioner",
            Description = "You have a specific target you must get lynched to win."
        };
        executioner.Abilities.Add(basicVote);
        executioner.Abilities.Add(new Ability { Name = "Target Elimination", Description = "Win if your assigned target is lynched. You are immune to night kills until your target dies." });

        // Add all created roles to the repository
        _roles.Add(townie);
        _roles.Add(doctor);
        _roles.Add(investigator);
        _roles.Add(vigilante);
        _roles.Add(mafioso);
        _roles.Add(godfather);
        _roles.Add(jester);
        _roles.Add(executioner);
    }

    public async Task<Result<Role>> GetFromIdAsync(int id)
    {
        // Simulate async
        return await Task.Run(() =>
        {
            Role? role = _roles.FirstOrDefault(r => r.Id == id);

            if (role == null)
            {
                var exception = new NotFoundException($"Role with id {id} not found.");
                return new Result<Role>(exception);
            }
            return role;
        });
    }
}