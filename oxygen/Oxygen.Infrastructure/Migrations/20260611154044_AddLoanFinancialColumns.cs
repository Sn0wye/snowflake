using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Oxygen.Infrastructure.Migrations
{
    /// <inheritdoc />
    public partial class AddLoanFinancialColumns : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AddColumn<decimal>(
                name: "interest_rate",
                table: "loan_applications",
                type: "numeric",
                nullable: false,
                defaultValue: 0m);

            migrationBuilder.AddColumn<decimal>(
                name: "monthly_payment",
                table: "loan_applications",
                type: "numeric",
                nullable: false,
                defaultValue: 0m);

            migrationBuilder.AddColumn<decimal>(
                name: "total_payment",
                table: "loan_applications",
                type: "numeric",
                nullable: false,
                defaultValue: 0m);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "interest_rate",
                table: "loan_applications");

            migrationBuilder.DropColumn(
                name: "monthly_payment",
                table: "loan_applications");

            migrationBuilder.DropColumn(
                name: "total_payment",
                table: "loan_applications");
        }
    }
}
