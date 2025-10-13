using System;
using Microsoft.EntityFrameworkCore.Migrations;
using Npgsql.EntityFrameworkCore.PostgreSQL.Metadata;

#nullable disable

#pragma warning disable CA1814 // Prefer jagged arrays over multidimensional

namespace API.DataAccess.Migrations
{
    /// <inheritdoc />
    public partial class Initial : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateTable(
                name: "Rulesets",
                columns: table => new
                {
                    Id = table.Column<int>(type: "integer", nullable: false)
                        .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                    Name = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                    Description = table.Column<string>(type: "character varying(512)", maxLength: 512, nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Rulesets", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "Abilities",
                columns: table => new
                {
                    Id = table.Column<int>(type: "integer", nullable: false)
                        .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                    Name = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                    Description = table.Column<string>(type: "text", nullable: false),
                    RulesetId = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Abilities", x => x.Id);
                    table.ForeignKey(
                        name: "FK_Abilities_Rulesets_RulesetId",
                        column: x => x.RulesetId,
                        principalTable: "Rulesets",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "GameSessions",
                columns: table => new
                {
                    Id = table.Column<int>(type: "integer", nullable: false)
                        .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                    RulesetId = table.Column<int>(type: "integer", nullable: false),
                    Status = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_GameSessions", x => x.Id);
                    table.ForeignKey(
                        name: "FK_GameSessions_Rulesets_RulesetId",
                        column: x => x.RulesetId,
                        principalTable: "Rulesets",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "Roles",
                columns: table => new
                {
                    Id = table.Column<int>(type: "integer", nullable: false)
                        .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                    Name = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false),
                    Description = table.Column<string>(type: "character varying(512)", maxLength: 512, nullable: false),
                    RulesetId = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Roles", x => x.Id);
                    table.ForeignKey(
                        name: "FK_Roles_Rulesets_RulesetId",
                        column: x => x.RulesetId,
                        principalTable: "Rulesets",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "Rounds",
                columns: table => new
                {
                    Id = table.Column<int>(type: "integer", nullable: false)
                        .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                    StartTime = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                    EndTime = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                    RoundNumber = table.Column<int>(type: "integer", nullable: false),
                    GameId = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Rounds", x => x.Id);
                    table.ForeignKey(
                        name: "FK_Rounds_GameSessions_GameId",
                        column: x => x.GameId,
                        principalTable: "GameSessions",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "AbilityRole",
                columns: table => new
                {
                    AbilitiesId = table.Column<int>(type: "integer", nullable: false),
                    AssociatedRolesId = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_AbilityRole", x => new { x.AbilitiesId, x.AssociatedRolesId });
                    table.ForeignKey(
                        name: "FK_AbilityRole_Abilities_AbilitiesId",
                        column: x => x.AbilitiesId,
                        principalTable: "Abilities",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                    table.ForeignKey(
                        name: "FK_AbilityRole_Roles_AssociatedRolesId",
                        column: x => x.AssociatedRolesId,
                        principalTable: "Roles",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "Players",
                columns: table => new
                {
                    Id = table.Column<int>(type: "integer", nullable: false)
                        .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                    Name = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                    RoleId = table.Column<int>(type: "integer", nullable: true),
                    GameId = table.Column<int>(type: "integer", nullable: true),
                    IsEliminated = table.Column<bool>(type: "boolean", nullable: false),
                    GameSessionId = table.Column<int>(type: "integer", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Players", x => x.Id);
                    table.ForeignKey(
                        name: "FK_Players_GameSessions_GameId",
                        column: x => x.GameId,
                        principalTable: "GameSessions",
                        principalColumn: "Id");
                    table.ForeignKey(
                        name: "FK_Players_GameSessions_GameSessionId",
                        column: x => x.GameSessionId,
                        principalTable: "GameSessions",
                        principalColumn: "Id");
                    table.ForeignKey(
                        name: "FK_Players_Roles_RoleId",
                        column: x => x.RoleId,
                        principalTable: "Roles",
                        principalColumn: "Id");
                });

            migrationBuilder.CreateTable(
                name: "RoleVisibility",
                columns: table => new
                {
                    CanBeSeenById = table.Column<int>(type: "integer", nullable: false),
                    CanSeeId = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RoleVisibility", x => new { x.CanBeSeenById, x.CanSeeId });
                    table.ForeignKey(
                        name: "FK_RoleVisibility_Roles_CanBeSeenById",
                        column: x => x.CanBeSeenById,
                        principalTable: "Roles",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                    table.ForeignKey(
                        name: "FK_RoleVisibility_Roles_CanSeeId",
                        column: x => x.CanSeeId,
                        principalTable: "Roles",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "PlayerVisibility",
                columns: table => new
                {
                    CanBeSeenById = table.Column<int>(type: "integer", nullable: false),
                    CanSeeId = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PlayerVisibility", x => new { x.CanBeSeenById, x.CanSeeId });
                    table.ForeignKey(
                        name: "FK_PlayerVisibility_Players_CanBeSeenById",
                        column: x => x.CanBeSeenById,
                        principalTable: "Players",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                    table.ForeignKey(
                        name: "FK_PlayerVisibility_Players_CanSeeId",
                        column: x => x.CanSeeId,
                        principalTable: "Players",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.InsertData(
                table: "Rulesets",
                columns: new[] { "Id", "Description", "Name" },
                values: new object[] { 1, "Town members aim to find and eliminate all evil roles, while Mafia and Neutral roles have their own secret goals, often involving the elimination of Town members. The game alternates between night, where players use their unique abilities, and day, where they discuss information, share their 'wills', and vote to hang someone.", "Town of Salem" });

            migrationBuilder.InsertData(
                table: "Abilities",
                columns: new[] { "Id", "Description", "Name", "RulesetId" },
                values: new object[,]
                {
                    { 1, "Participate in daily voting to lynch a suspect.", "Vote", 1 },
                    { 3, "Can defend themselves against night attacks.", "Defense", 1 },
                    { 4, "Choose one player to protect from death each night.", "Heal", 1 },
                    { 5, "Choose one player to investigate each night and learn their role category.", "Investigate", 1 },
                    { 6, "Choose one player to kill each night. Limited uses.", "Shoot", 1 },
                    { 7, "Execute the Mafia's chosen target at night.", "Mafia Kill", 1 },
                    { 8, "Choose the Mafia's nightly target. Appears as Townie to investigators.", "Organize Kill", 1 },
                    { 9, "If lynched, you will kill one player who voted for you that night.", "Trick", 1 },
                    { 10, "Win if your assigned target is lynched. You are immune to night kills until your target dies.", "Target Elimination", 1 }
                });

            migrationBuilder.InsertData(
                table: "GameSessions",
                columns: new[] { "Id", "RulesetId", "Status" },
                values: new object[] { 1, 1, 100 });

            migrationBuilder.InsertData(
                table: "Roles",
                columns: new[] { "Id", "Description", "Name", "RulesetId" },
                values: new object[,]
                {
                    { 1, "A regular citizen of the town. Your goal is to eliminate all threats.", "Townie", 1 },
                    { 2, "You are a medical professional dedicated to saving lives.", "Doctor", 1 },
                    { 3, "You seek the truth and uncover secrets hidden in the town.", "Investigator", 1 },
                    { 4, "You take justice into your own hands, even if it means getting your hands dirty.", "Vigilante", 1 },
                    { 5, "A loyal member of the Mafia. You carry out the family's nightly kills.", "Mafioso", 1 },
                    { 6, "The cunning leader of the Mafia. You are immune to basic investigations.", "Godfather", 1 },
                    { 7, "Your only goal is to be lynched by the town.", "Jester", 1 },
                    { 8, "You have a specific target you must get lynched to win.", "Executioner", 1 }
                });

            migrationBuilder.InsertData(
                table: "Players",
                columns: new[] { "Id", "GameId", "GameSessionId", "IsEliminated", "Name", "RoleId" },
                values: new object[,]
                {
                    { 1, 1, null, false, "Alice Townie", 1 },
                    { 2, 1, null, false, "John Doctor", 2 },
                    { 3, 1, null, false, "Emily Investigator", 3 },
                    { 4, 1, null, false, "Michael Vigilante", 4 },
                    { 5, 1, null, false, "Sarah Mafioso", 5 },
                    { 6, 1, null, false, "Jessica Godfather", 6 },
                    { 7, 1, null, false, "David Jester", 7 },
                    { 8, 1, null, false, "Ashley Executioner", 8 },
                    { 9, 1, null, false, "Matthew Townie", 1 },
                    { 10, 1, null, false, "Amanda Doctor", 2 },
                    { 11, 1, null, false, "Joshua Investigator", 3 },
                    { 12, 1, null, false, "Jennifer Vigilante", 4 },
                    { 13, 1, null, false, "Daniel Mafioso", 5 },
                    { 14, 1, null, false, "Elizabeth Godfather", 6 },
                    { 15, 1, null, false, "James Jester", 7 },
                    { 16, 1, null, false, "Charlie Executioner", 8 },
                    { 17, 1, null, false, "Kyle Townie", 1 },
                    { 18, 1, null, false, "Bob Doctor", 2 },
                    { 19, 1, null, false, "Megan Investigator", 3 },
                    { 20, 1, null, false, "Laura Vigilante", 4 }
                });

            migrationBuilder.CreateIndex(
                name: "IX_Abilities_RulesetId",
                table: "Abilities",
                column: "RulesetId");

            migrationBuilder.CreateIndex(
                name: "IX_AbilityRole_AssociatedRolesId",
                table: "AbilityRole",
                column: "AssociatedRolesId");

            migrationBuilder.CreateIndex(
                name: "IX_GameSessions_RulesetId",
                table: "GameSessions",
                column: "RulesetId");

            migrationBuilder.CreateIndex(
                name: "IX_Players_GameId",
                table: "Players",
                column: "GameId");

            migrationBuilder.CreateIndex(
                name: "IX_Players_GameSessionId",
                table: "Players",
                column: "GameSessionId");

            migrationBuilder.CreateIndex(
                name: "IX_Players_RoleId",
                table: "Players",
                column: "RoleId");

            migrationBuilder.CreateIndex(
                name: "IX_PlayerVisibility_CanSeeId",
                table: "PlayerVisibility",
                column: "CanSeeId");

            migrationBuilder.CreateIndex(
                name: "IX_Roles_RulesetId",
                table: "Roles",
                column: "RulesetId");

            migrationBuilder.CreateIndex(
                name: "IX_RoleVisibility_CanSeeId",
                table: "RoleVisibility",
                column: "CanSeeId");

            migrationBuilder.CreateIndex(
                name: "IX_Rounds_GameId",
                table: "Rounds",
                column: "GameId");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "AbilityRole");

            migrationBuilder.DropTable(
                name: "PlayerVisibility");

            migrationBuilder.DropTable(
                name: "RoleVisibility");

            migrationBuilder.DropTable(
                name: "Rounds");

            migrationBuilder.DropTable(
                name: "Abilities");

            migrationBuilder.DropTable(
                name: "Players");

            migrationBuilder.DropTable(
                name: "GameSessions");

            migrationBuilder.DropTable(
                name: "Roles");

            migrationBuilder.DropTable(
                name: "Rulesets");
        }
    }
}
