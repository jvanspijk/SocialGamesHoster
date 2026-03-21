using System;
using System.Net;
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
                    Name = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
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
                name: "RoundPhases",
                columns: table => new
                {
                    Id = table.Column<int>(type: "integer", nullable: false)
                        .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                    RulesetId = table.Column<int>(type: "integer", nullable: false),
                    Name = table.Column<string>(type: "text", nullable: false),
                    Description = table.Column<string>(type: "text", nullable: false),
                    Order = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RoundPhases", x => x.Id);
                    table.ForeignKey(
                        name: "FK_RoundPhases_Rulesets_RulesetId",
                        column: x => x.RulesetId,
                        principalTable: "Rulesets",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "RoleAbilities",
                columns: table => new
                {
                    RoleId = table.Column<int>(type: "integer", nullable: false),
                    AbilityId = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RoleAbilities", x => new { x.RoleId, x.AbilityId });
                    table.ForeignKey(
                        name: "FK_RoleAbilities_Abilities_AbilityId",
                        column: x => x.AbilityId,
                        principalTable: "Abilities",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                    table.ForeignKey(
                        name: "FK_RoleAbilities_Roles_RoleId",
                        column: x => x.RoleId,
                        principalTable: "Roles",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "RoleKnowledges",
                columns: table => new
                {
                    SourceId = table.Column<int>(type: "integer", nullable: false),
                    TargetId = table.Column<int>(type: "integer", nullable: false),
                    KnowledgeType = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RoleKnowledges", x => new { x.SourceId, x.TargetId });
                    table.ForeignKey(
                        name: "FK_RoleKnowledges_Roles_SourceId",
                        column: x => x.SourceId,
                        principalTable: "Roles",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                    table.ForeignKey(
                        name: "FK_RoleKnowledges_Roles_TargetId",
                        column: x => x.TargetId,
                        principalTable: "Roles",
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
                    RoundNumber = table.Column<int>(type: "integer", nullable: false),
                    RoundStartedAt = table.Column<DateTimeOffset>(type: "timestamp with time zone", nullable: false),
                    CurrentPhaseId = table.Column<int>(type: "integer", nullable: true),
                    PhaseStartedAt = table.Column<DateTimeOffset>(type: "timestamp with time zone", nullable: false),
                    Status = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_GameSessions", x => x.Id);
                    table.ForeignKey(
                        name: "FK_GameSessions_RoundPhases_CurrentPhaseId",
                        column: x => x.CurrentPhaseId,
                        principalTable: "RoundPhases",
                        principalColumn: "Id");
                    table.ForeignKey(
                        name: "FK_GameSessions_Rulesets_RulesetId",
                        column: x => x.RulesetId,
                        principalTable: "Rulesets",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "ChatChannels",
                columns: table => new
                {
                    Id = table.Column<int>(type: "integer", nullable: false)
                        .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn),
                    Name = table.Column<string>(type: "text", nullable: false),
                    GameId = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_ChatChannels", x => x.Id);
                    table.ForeignKey(
                        name: "FK_ChatChannels_GameSessions_GameId",
                        column: x => x.GameId,
                        principalTable: "GameSessions",
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
                    IP = table.Column<IPAddress>(type: "inet", nullable: true),
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
                name: "ChatChannelMembership",
                columns: table => new
                {
                    PlayerId = table.Column<int>(type: "integer", nullable: false),
                    ChannelId = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_ChatChannelMembership", x => new { x.PlayerId, x.ChannelId });
                    table.ForeignKey(
                        name: "FK_ChatChannelMembership_ChatChannels_ChannelId",
                        column: x => x.ChannelId,
                        principalTable: "ChatChannels",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                    table.ForeignKey(
                        name: "FK_ChatChannelMembership_Players_PlayerId",
                        column: x => x.PlayerId,
                        principalTable: "Players",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateTable(
                name: "ChatMessages",
                columns: table => new
                {
                    Id = table.Column<Guid>(type: "uuid", nullable: false),
                    Content = table.Column<string>(type: "text", nullable: false),
                    SentAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                    SenderId = table.Column<int>(type: "integer", nullable: false),
                    ChannelId = table.Column<int>(type: "integer", nullable: false),
                    IsDeleted = table.Column<bool>(type: "boolean", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_ChatMessages", x => x.Id);
                    table.ForeignKey(
                        name: "FK_ChatMessages_ChatChannels_ChannelId",
                        column: x => x.ChannelId,
                        principalTable: "ChatChannels",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                    table.ForeignKey(
                        name: "FK_ChatMessages_Players_SenderId",
                        column: x => x.SenderId,
                        principalTable: "Players",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Restrict);
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
                values: new object[,]
                {
                    { 1, "Town members aim to find and eliminate all evil roles, while Mafia and Neutral roles have their own secret goals, often involving the elimination of Town members. The game alternates between night, where players use their unique abilities, and day, where they discuss information, share their 'wills', and vote to hang someone.", "Town of Salem" },
                    { 2, "Blackjack, also known as 21, is a popular card game where players aim to have a hand value as close to 21 as possible without exceeding it. Players compete against the dealer rather than each other. The game involves strategic decisions such as hitting, standing, doubling down, and splitting pairs.", "Blackjack" }
                });

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
                    { 10, "Win if your assigned target is lynched. You are immune to night kills until your target dies.", "Target Elimination", 1 },
                    { 100, "Request an additional card to increase hand value.", "Hit", 2 },
                    { 101, "Keep current hand and end turn.", "Stand", 2 },
                    { 102, "Double the initial bet and receive one final card.", "Double Down", 2 },
                    { 103, "If dealt a pair, split into two separate hands.", "Split", 2 }
                });

            migrationBuilder.InsertData(
                table: "GameSessions",
                columns: new[] { "Id", "CurrentPhaseId", "PhaseStartedAt", "RoundNumber", "RoundStartedAt", "RulesetId", "Status" },
                values: new object[] { 1, null, new DateTimeOffset(new DateTime(2026, 1, 1, 0, 0, 0, 0, DateTimeKind.Unspecified), new TimeSpan(0, 0, 0, 0, 0)), 0, new DateTimeOffset(new DateTime(2026, 1, 1, 0, 0, 0, 0, DateTimeKind.Unspecified), new TimeSpan(0, 0, 0, 0, 0)), 1, 100 });

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
                    { 8, "You have a specific target you must get lynched to win.", "Executioner", 1 },
                    { 100, "A participant in the game aiming to beat the dealer by achieving a hand value of 21 or less.", "Player", 2 },
                    { 101, "The house representative who manages the game, deals cards, and plays against the players.", "Dealer", 2 }
                });

            migrationBuilder.InsertData(
                table: "ChatChannels",
                columns: new[] { "Id", "GameId", "Name" },
                values: new object[] { 1, 1, "Global" });

            migrationBuilder.InsertData(
                table: "Players",
                columns: new[] { "Id", "GameId", "GameSessionId", "IP", "IsEliminated", "Name", "RoleId" },
                values: new object[,]
                {
                    { 1, 1, null, null, false, "Alice Townie", 1 },
                    { 2, 1, null, null, false, "John Doctor", 2 },
                    { 3, 1, null, null, false, "Emily Investigator", 3 },
                    { 4, 1, null, null, false, "Michael Vigilante", 4 },
                    { 5, 1, null, null, false, "Sarah Mafioso", 5 },
                    { 6, 1, null, null, false, "Jessica Godfather", 6 },
                    { 7, 1, null, null, false, "David Jester", 7 },
                    { 8, 1, null, null, false, "Ashley Executioner", 8 },
                    { 9, 1, null, null, false, "Matthew Townie", 1 },
                    { 10, 1, null, null, false, "Amanda Doctor", 2 },
                    { 11, 1, null, null, false, "Joshua Investigator", 3 },
                    { 12, 1, null, null, false, "Jennifer Vigilante", 4 },
                    { 13, 1, null, null, false, "Daniel Mafioso", 5 },
                    { 14, 1, null, null, false, "Elizabeth Godfather", 6 },
                    { 15, 1, null, null, false, "James Jester", 7 },
                    { 16, 1, null, null, false, "Charlie Executioner", 8 },
                    { 17, 1, null, null, false, "Kyle Townie", 1 },
                    { 18, 1, null, null, false, "Bob Doctor", 2 },
                    { 19, 1, null, null, false, "Megan Investigator", 3 },
                    { 20, 1, null, null, false, "Laura Vigilante", 4 }
                });

            migrationBuilder.InsertData(
                table: "RoleAbilities",
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
                    { 10, 8 },
                    { 100, 100 },
                    { 101, 100 },
                    { 102, 100 },
                    { 103, 100 },
                    { 100, 101 },
                    { 101, 101 }
                });

            migrationBuilder.InsertData(
                table: "RoleKnowledges",
                columns: new[] { "SourceId", "TargetId", "KnowledgeType" },
                values: new object[,]
                {
                    { 5, 6, 100 },
                    { 6, 5, 100 },
                    { 100, 100, 100 },
                    { 100, 101, 100 },
                    { 101, 100, 100 },
                    { 101, 101, 100 }
                });

            migrationBuilder.CreateIndex(
                name: "IX_Abilities_RulesetId",
                table: "Abilities",
                column: "RulesetId");

            migrationBuilder.CreateIndex(
                name: "IX_ChatChannelMembership_ChannelId",
                table: "ChatChannelMembership",
                column: "ChannelId");

            migrationBuilder.CreateIndex(
                name: "IX_ChatChannels_GameId",
                table: "ChatChannels",
                column: "GameId");

            migrationBuilder.CreateIndex(
                name: "IX_ChatMessage_Channel_Time",
                table: "ChatMessages",
                columns: new[] { "ChannelId", "SentAt" });

            migrationBuilder.CreateIndex(
                name: "IX_ChatMessages_SenderId",
                table: "ChatMessages",
                column: "SenderId");

            migrationBuilder.CreateIndex(
                name: "IX_GameSessions_CurrentPhaseId",
                table: "GameSessions",
                column: "CurrentPhaseId");

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
                name: "IX_RoleAbilities_AbilityId",
                table: "RoleAbilities",
                column: "AbilityId");

            migrationBuilder.CreateIndex(
                name: "IX_RoleKnowledges_TargetId",
                table: "RoleKnowledges",
                column: "TargetId");

            migrationBuilder.CreateIndex(
                name: "IX_Roles_RulesetId",
                table: "Roles",
                column: "RulesetId");

            migrationBuilder.CreateIndex(
                name: "IX_RoundPhases_RulesetId",
                table: "RoundPhases",
                column: "RulesetId");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "ChatChannelMembership");

            migrationBuilder.DropTable(
                name: "ChatMessages");

            migrationBuilder.DropTable(
                name: "PlayerVisibility");

            migrationBuilder.DropTable(
                name: "RoleAbilities");

            migrationBuilder.DropTable(
                name: "RoleKnowledges");

            migrationBuilder.DropTable(
                name: "ChatChannels");

            migrationBuilder.DropTable(
                name: "Players");

            migrationBuilder.DropTable(
                name: "Abilities");

            migrationBuilder.DropTable(
                name: "GameSessions");

            migrationBuilder.DropTable(
                name: "Roles");

            migrationBuilder.DropTable(
                name: "RoundPhases");

            migrationBuilder.DropTable(
                name: "Rulesets");
        }
    }
}
