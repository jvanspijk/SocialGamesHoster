using API.DataAccess;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Design;
using Microsoft.Extensions.Configuration;

namespace API;

public class APIDatabaseContextFactory : IDesignTimeDbContextFactory<APIDatabaseContext>
{
    public APIDatabaseContext CreateDbContext(string[] args)
    {
        var propertiesPath = Path.Combine(Directory.GetCurrentDirectory(), "Properties");
        IConfigurationRoot configuration = new ConfigurationBuilder()
            .SetBasePath(propertiesPath)
            .AddJsonFile("appsettings.json", optional: true)
            .AddJsonFile("appsettings.Development.json", optional: true)
            .Build();

        var optionsBuilder = new DbContextOptionsBuilder<APIDatabaseContext>();
        var connectionString = configuration.GetConnectionString("DefaultConnection");
        optionsBuilder.UseNpgsql(connectionString);

        return new APIDatabaseContext(optionsBuilder.Options);
    }
}
