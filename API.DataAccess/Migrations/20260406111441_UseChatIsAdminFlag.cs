using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace API.DataAccess.Migrations
{
    /// <inheritdoc />
    public partial class UseChatIsAdminFlag : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AddColumn<bool>(
                name: "IsAdmin",
                table: "ChatMessages",
                type: "boolean",
                nullable: false,
                defaultValue: false);

            migrationBuilder.Sql(
                "UPDATE \"ChatMessages\" SET \"IsAdmin\" = TRUE WHERE \"SenderType\" = 1");

            migrationBuilder.DropColumn(
                name: "SenderType",
                table: "ChatMessages");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "IsAdmin",
                table: "ChatMessages");

            migrationBuilder.AddColumn<int>(
                name: "SenderType",
                table: "ChatMessages",
                type: "integer",
                nullable: false,
                defaultValue: 0);
        }
    }
}
