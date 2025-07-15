using API.Models;
using Microsoft.EntityFrameworkCore;
namespace API.DataAccess;

public class APIDatabaseContext : DbContext
{
    public APIDatabaseContext(DbContextOptions options) : base(options) { }

    protected override void OnModelCreating(ModelBuilder builder)
    {
        base.OnModelCreating(builder);
    }
}
