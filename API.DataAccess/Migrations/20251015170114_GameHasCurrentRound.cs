using Microsoft.EntityFrameworkCore.Migrations;
using Npgsql.EntityFrameworkCore.PostgreSQL.Metadata;

#nullable disable

#pragma warning disable CA1814 // Prefer jagged arrays over multidimensional

namespace API.DataAccess.Migrations
{
    /// <inheritdoc />
    public partial class GameHasCurrentRound : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "AbilityRole");

            migrationBuilder.DropTable(
                name: "RoleVisibility");

            migrationBuilder.AlterColumn<int>(
                name: "Id",
                table: "GameSessions",
                type: "integer",
                nullable: false,
                oldClrType: typeof(int),
                oldType: "integer")
                .OldAnnotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn);

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
                    { 10, 8 }
                });

            migrationBuilder.InsertData(
                table: "RoleKnowledges",
                columns: new[] { "SourceId", "TargetId", "KnowledgeType" },
                values: new object[,]
                {
                    { 5, 6, 100 },
                    { 6, 5, 100 }
                });

            migrationBuilder.CreateIndex(
                name: "IX_RoleAbilities_AbilityId",
                table: "RoleAbilities",
                column: "AbilityId");

            migrationBuilder.CreateIndex(
                name: "IX_RoleKnowledges_TargetId",
                table: "RoleKnowledges",
                column: "TargetId");

            migrationBuilder.AddForeignKey(
                name: "FK_GameSessions_Rounds_Id",
                table: "GameSessions",
                column: "Id",
                principalTable: "Rounds",
                principalColumn: "Id");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropForeignKey(
                name: "FK_GameSessions_Rounds_Id",
                table: "GameSessions");

            migrationBuilder.DropTable(
                name: "RoleAbilities");

            migrationBuilder.DropTable(
                name: "RoleKnowledges");

            migrationBuilder.AlterColumn<int>(
                name: "Id",
                table: "GameSessions",
                type: "integer",
                nullable: false,
                oldClrType: typeof(int),
                oldType: "integer")
                .Annotation("Npgsql:ValueGenerationStrategy", NpgsqlValueGenerationStrategy.IdentityByDefaultColumn);

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

            migrationBuilder.CreateIndex(
                name: "IX_AbilityRole_AssociatedRolesId",
                table: "AbilityRole",
                column: "AssociatedRolesId");

            migrationBuilder.CreateIndex(
                name: "IX_RoleVisibility_CanSeeId",
                table: "RoleVisibility",
                column: "CanSeeId");
        }
    }
}
