using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Diagnostics;
using TenantlyNotificationService.Models.Interfaces;

namespace TenantlyNotificationService.Data.Interceptors;

public class AuditInterceptor : SaveChangesInterceptor
{
    private readonly ILogger<AuditInterceptor>? _logger;

    public AuditInterceptor(ILogger<AuditInterceptor>? logger = null)
    {
        _logger = logger;
    }

    public override InterceptionResult<int> SavingChanges(DbContextEventData eventData, InterceptionResult<int> result)
    {
        UpdateTimestamps(eventData.Context);
        return base.SavingChanges(eventData, result);
    }

    public override ValueTask<InterceptionResult<int>> SavingChangesAsync(DbContextEventData eventData, InterceptionResult<int> result, CancellationToken cancellationToken = default)
    {
        UpdateTimestamps(eventData.Context);
        return base.SavingChangesAsync(eventData, result, cancellationToken);
    }

    private void UpdateTimestamps(DbContext? context)
    {
        if (context == null) return;

        var utcNow = DateTime.UtcNow;
        var changedEntries = 0;

        var entries = context.ChangeTracker.Entries<IAuditableEntity>()
            .Where(e => e.State == EntityState.Added || e.State == EntityState.Modified)
            .ToList();

        foreach (var entry in entries)
        {
            var entityName = entry.Entity.GetType().Name;

            // Handle CreatedAt for new entities
            if (entry.State == EntityState.Added)
            {
                entry.Entity.CreatedAt = utcNow;
                entry.Entity.UpdatedAt = utcNow;
                changedEntries++;
                _logger?.LogDebug("Setting timestamps for new {EntityName}", entityName);
            }
            // Handle UpdatedAt for modified entities
            else if (entry.State == EntityState.Modified)
            {
                entry.Entity.UpdatedAt = utcNow;
                changedEntries++;
                _logger?.LogDebug("Updating timestamp for modified {EntityName}", entityName);
            }
        }

        if (changedEntries > 0)
        {
            _logger?.LogDebug("Updated timestamps for {Count} entities", changedEntries);
        }
    }
}