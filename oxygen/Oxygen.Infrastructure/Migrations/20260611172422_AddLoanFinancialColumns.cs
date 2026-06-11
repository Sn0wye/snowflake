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
            migrationBuilder.AlterColumn<decimal>(
                name: "amount",
                table: "loan_applications",
                type: "numeric(18,2)",
                nullable: false,
                oldClrType: typeof(double),
                oldType: "double precision");

            migrationBuilder.AddColumn<decimal>(
                name: "interest_rate",
                table: "loan_applications",
                type: "numeric(18,6)",
                nullable: true);

            migrationBuilder.AddColumn<decimal>(
                name: "monthly_payment",
                table: "loan_applications",
                type: "numeric(18,2)",
                nullable: true);

            migrationBuilder.AddColumn<decimal>(
                name: "total_payment",
                table: "loan_applications",
                type: "numeric(18,2)",
                nullable: true);
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

            migrationBuilder.AlterColumn<double>(
                name: "amount",
                table: "loan_applications",
                type: "double precision",
                nullable: false,
                oldClrType: typeof(decimal),
                oldType: "numeric");
        }
    }
}
