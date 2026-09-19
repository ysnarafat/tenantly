using Microsoft.Extensions.Configuration;

namespace TenantlyNotificationService.Configuration;

public static class ConnectionStringResolver
{
    public static string Resolve(IConfiguration configuration)
    {
        return configuration.GetConnectionString("DefaultConnection")
            ?? configuration.GetSection("Database:ConnectionString").Value
            ?? throw new InvalidOperationException("Connection string not found.");
    }
}
