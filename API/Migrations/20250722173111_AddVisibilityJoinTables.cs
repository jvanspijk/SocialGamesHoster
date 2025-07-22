using System.Collections.Generic;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

#pragma warning disable CA1814 // Prefer jagged arrays over multidimensional

namespace API.Migrations
{
    /// <inheritdoc />
    public partial class AddVisibilityJoinTables : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "RolesVisibleToRole",
                table: "Roles");

            migrationBuilder.DropColumn(
                name: "PlayersVisibleToPlayer",
                table: "Players");

            migrationBuilder.CreateTable(
                name: "PlayerVisibility",
                columns: table => new
                {
                    PlayerId = table.Column<int>(type: "integer", nullable: false),
                    VisiblePlayerId = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PlayerVisibility", x => new { x.PlayerId, x.VisiblePlayerId });
                    table.ForeignKey(
                        name: "FK_PlayerVisibility_Players_PlayerId",
                        column: x => x.PlayerId,
                        principalTable: "Players",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Restrict);
                    table.ForeignKey(
                        name: "FK_PlayerVisibility_Players_VisiblePlayerId",
                        column: x => x.VisiblePlayerId,
                        principalTable: "Players",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Restrict);
                });

            migrationBuilder.CreateTable(
                name: "RoleVisibility",
                columns: table => new
                {
                    RoleId = table.Column<int>(type: "integer", nullable: false),
                    VisibleRoleId = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_RoleVisibility", x => new { x.RoleId, x.VisibleRoleId });
                    table.ForeignKey(
                        name: "FK_RoleVisibility_Roles_RoleId",
                        column: x => x.RoleId,
                        principalTable: "Roles",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Restrict);
                    table.ForeignKey(
                        name: "FK_RoleVisibility_Roles_VisibleRoleId",
                        column: x => x.VisibleRoleId,
                        principalTable: "Roles",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Restrict);
                });

            migrationBuilder.InsertData(
                table: "RoleVisibility",
                columns: new[] { "RoleId", "VisibleRoleId" },
                values: new object[,]
                {
                    { 5, 6 },
                    { 6, 5 }
                });

            migrationBuilder.CreateIndex(
                name: "IX_PlayerVisibility_VisiblePlayerId",
                table: "PlayerVisibility",
                column: "VisiblePlayerId");

            migrationBuilder.CreateIndex(
                name: "IX_RoleVisibility_VisibleRoleId",
                table: "RoleVisibility",
                column: "VisibleRoleId");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "PlayerVisibility");

            migrationBuilder.DropTable(
                name: "RoleVisibility");

            migrationBuilder.AddColumn<List<int>>(
                name: "RolesVisibleToRole",
                table: "Roles",
                type: "integer[]",
                nullable: false);

            migrationBuilder.AddColumn<List<int>>(
                name: "PlayersVisibleToPlayer",
                table: "Players",
                type: "integer[]",
                nullable: false);

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 1,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 2,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 3,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 4,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 5,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 6,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 7,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 8,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 9,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 10,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 11,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 12,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 13,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 14,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 15,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 16,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 17,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 18,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 19,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Players",
                keyColumn: "Id",
                keyValue: 20,
                column: "PlayersVisibleToPlayer",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Roles",
                keyColumn: "Id",
                keyValue: 1,
                column: "RolesVisibleToRole",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Roles",
                keyColumn: "Id",
                keyValue: 2,
                column: "RolesVisibleToRole",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Roles",
                keyColumn: "Id",
                keyValue: 3,
                column: "RolesVisibleToRole",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Roles",
                keyColumn: "Id",
                keyValue: 4,
                column: "RolesVisibleToRole",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Roles",
                keyColumn: "Id",
                keyValue: 5,
                column: "RolesVisibleToRole",
                value: new List<int> { 6, 5 });

            migrationBuilder.UpdateData(
                table: "Roles",
                keyColumn: "Id",
                keyValue: 6,
                column: "RolesVisibleToRole",
                value: new List<int> { 6, 5 });

            migrationBuilder.UpdateData(
                table: "Roles",
                keyColumn: "Id",
                keyValue: 7,
                column: "RolesVisibleToRole",
                value: new List<int>());

            migrationBuilder.UpdateData(
                table: "Roles",
                keyColumn: "Id",
                keyValue: 8,
                column: "RolesVisibleToRole",
                value: new List<int>());
        }
    }
}
