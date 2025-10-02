using System.Collections.Generic;
using Microsoft.EntityFrameworkCore.Migrations;
using Npgsql.EntityFrameworkCore.PostgreSQL.Metadata;

#nullable disable

#pragma warning disable CA1814 // Prefer jagged arrays over multidimensional

namespace API.Migrations
{
    /// <inheritdoc />
    public partial class InitialCreate : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateTable(
                name: "Abilities",
                columns: table => new
                {
                    Id = table.Column<int>(type: "integer", nullable: false)
                        .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                    Name = table.Column<string>(type: "text", nullable: false),
                    Description = table.Column<string>(type: "text", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Abilities", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "Roles",
                columns: table => new
                {
                    Id = table.Column<int>(type: "integer", nullable: false)
                        .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                    Name = table.Column<string>(type: "text", nullable: false),
                    Description = table.Column<string>(type: "text", nullable: true),
                    RolesVisibleToRole = table.Column<int[]>(type: "integer[]", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Roles", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "Players",
                columns: table => new
                {
                    Id = table.Column<int>(type: "integer", nullable: false)
                        .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                    Name = table.Column<string>(type: "text", nullable: false),
                    RoleId = table.Column<int>(type: "integer", nullable: true),
                    PlayersVisibleToPlayer = table.Column<List<int>>(type: "integer[]", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Players", x => x.Id);
                    table.ForeignKey(
                        name: "FK_Players_Roles_RoleId",
                        column: x => x.RoleId,
                        principalTable: "Roles",
                        principalColumn: "Id");
                });

            migrationBuilder.CreateTable(
                name: "RoleAbilityAssociations",
                columns: table => new
                {
                    RoleId = table.Column<int>(type: "integer", nullable: false),
                    AbilityId = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RoleAbilityAssociations", x => new { x.RoleId, x.AbilityId });
                    table.ForeignKey(
                        name: "FK_RoleAbilityAssociations_Abilities_AbilityId",
                        column: x => x.AbilityId,
                        principalTable: "Abilities",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                    table.ForeignKey(
                        name: "FK_RoleAbilityAssociations_Roles_RoleId",
                        column: x => x.RoleId,
                        principalTable: "Roles",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.InsertData(
                table: "Abilities",
                columns: new[] { "Id", "Description", "Name" },
                values: new object[,]
                {
                    { 1, "Participate in daily voting to lynch a suspect.", "Vote" },
                    { 3, "Can defend themselves against night attacks.", "Defense" },
                    { 4, "Choose one player to protect from death each night.", "Heal" },
                    { 5, "Choose one player to investigate each night and learn their role category.", "Investigate" },
                    { 6, "Choose one player to kill each night. Limited uses.", "Shoot" },
                    { 7, "Execute the Mafia's chosen target at night.", "Mafia Kill" },
                    { 8, "Choose the Mafia's nightly target. Appears as Townie to investigators.", "Organize Kill" },
                    { 9, "If lynched, you will kill one player who voted for you that night.", "Trick" },
                    { 10, "Win if your assigned target is lynched. You are immune to night kills until your target dies.", "Target Elimination" }
                });

            migrationBuilder.InsertData(
                table: "Roles",
                columns: new[] { "Id", "Description", "Name", "RolesVisibleToRole" },
                values: new object[,]
                {
                    { 1, "A regular citizen of the town. Your goal is to eliminate all threats.", "Townie", new int[0] },
                    { 2, "You are a medical professional dedicated to saving lives.", "Doctor", new int[0] },
                    { 3, "You seek the truth and uncover secrets hidden in the town.", "Investigator", new int[0] },
                    { 4, "You take justice into your own hands, even if it means getting your hands dirty.", "Vigilante", new int[0] },
                    { 5, "A loyal member of the Mafia. You carry out the family's nightly kills.", "Mafioso", new[] { 6, 5 } },
                    { 6, "The cunning leader of the Mafia. You are immune to basic investigations.", "Godfather", new[] { 6, 5 } },
                    { 7, "Your only goal is to be lynched by the town.", "Jester", new int[0] },
                    { 8, "You have a specific target you must get lynched to win.", "Executioner", new int[0] }
                });

            migrationBuilder.InsertData(
                table: "Players",
                columns: new[] { "Id", "Name", "PlayersVisibleToPlayer", "RoleId" },
                values: new object[,]
                {
                    { 1, "Alice Townie", new List<int>(), 1 },
                    { 2, "John Doctor", new List<int>(), 2 },
                    { 3, "Emily Investigator", new List<int>(), 3 },
                    { 4, "Michael Vigilante", new List<int>(), 4 },
                    { 5, "Sarah Mafioso", new List<int>(), 5 },
                    { 6, "Jessica Godfather", new List<int>(), 6 },
                    { 7, "David Jester", new List<int>(), 7 },
                    { 8, "Ashley Executioner", new List<int>(), 8 },
                    { 9, "Matthew Townie", new List<int>(), 1 },
                    { 10, "Amanda Doctor", new List<int>(), 2 },
                    { 11, "Joshua Investigator", new List<int>(), 3 },
                    { 12, "Jennifer Vigilante", new List<int>(), 4 },
                    { 13, "Daniel Mafioso", new List<int>(), 5 },
                    { 14, "Elizabeth Godfather", new List<int>(), 6 },
                    { 15, "James Jester", new List<int>(), 7 },
                    { 16, "Charlie Executioner", new List<int>(), 8 },
                    { 17, "Kyle Townie", new List<int>(), 1 },
                    { 18, "Bob Doctor", new List<int>(), 2 },
                    { 19, "Megan Investigator", new List<int>(), 3 },
                    { 20, "Laura Vigilante", new List<int>(), 4 }
                });

            migrationBuilder.InsertData(
                table: "RoleAbilityAssociations",
                columns: new[] { "AbilityId", "RoleId" },
                values: new object[,]
                {
                    { 1, 1 },
                    { 1, 2 },
                    { 4, 2 },
                    { 1, 3 },
                    { 5, 3 },
                    { 1, 4 },
                    { 6, 4 },
                    { 1, 5 },
                    { 7, 5 },
                    { 1, 6 },
                    { 3, 6 },
                    { 8, 6 },
                    { 1, 7 },
                    { 9, 7 },
                    { 1, 8 },
                    { 10, 8 }
                });

            migrationBuilder.CreateIndex(
                name: "IX_Players_RoleId",
                table: "Players",
                column: "RoleId");

            migrationBuilder.CreateIndex(
                name: "IX_RoleAbilityAssociations_AbilityId",
                table: "RoleAbilityAssociations",
                column: "AbilityId");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "Players");

            migrationBuilder.DropTable(
                name: "RoleAbilityAssociations");

            migrationBuilder.DropTable(
                name: "Abilities");

            migrationBuilder.DropTable(
                name: "Roles");
        }
    }
}
