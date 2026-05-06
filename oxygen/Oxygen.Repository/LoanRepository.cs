using Oxygen.Infrastructure;

using Microsoft.EntityFrameworkCore;

namespace Oxygen.Repository;

public class LoanRepository(ApplicationDbContext context) : ILoanRepository
{
    public async Task<Domain.Entities.LoanApplication?> FindAsync(int id)
    {
        return await context.LoanApplications.FirstOrDefaultAsync(loan => loan.Id == id);
    }

    public async Task<Oxygen.Domain.Entities.LoanApplication> AddAsync(Oxygen.Domain.Entities.LoanApplication loanApplication)
    {
        await context.LoanApplications.AddAsync(loanApplication);
        await context.SaveChangesAsync();
        return loanApplication;
    }
}