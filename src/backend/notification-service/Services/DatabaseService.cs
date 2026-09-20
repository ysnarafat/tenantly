using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Options;
using TenantlyNotificationService.Configuration;
using TenantlyNotificationService.Data;
using TenantlyNotificationService.Models;
using TenantlyNotificationService.Models.Enums;

namespace TenantlyNotificationService.Services;

public class DatabaseService : IDatabaseService
{
    private readonly TenantlyDbContext _context;
    private readonly NotificationSettings _settings;
    private readonly ILogger<DatabaseService> _logger;

    public DatabaseService(TenantlyDbContext context, IOptions<NotificationSettings> settings, ILogger<DatabaseService> logger)
    {
        _context = context;
        _settings = settings.Value;
        _logger = logger;
    }

    public async Task<IEnumerable<NotificationQueue>> GetPendingNotificationsAsync()
    {
        try
        {
            return await _context.NotificationQueue
                .Where(n => n.Status == NotificationStatus.Pending.ToString() && n.RetryCount < _settings.MaxRetryAttempts)
                .OrderBy(n => n.CreatedAt)
                .ToListAsync();
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error retrieving pending notifications");
            return Enumerable.Empty<NotificationQueue>();
        }
    }

    public async Task UpdateNotificationStatusAsync(int notificationId, NotificationStatus status, string? errorMessage = null)
    {
        try
        {
            var notification = await _context.NotificationQueue.FindAsync(notificationId);
            if (notification != null)
            {
                notification.ErrorMessage = errorMessage;
                notification.UpdatedAt = DateTime.UtcNow;

                if (status == NotificationStatus.Failed)
                {
                    notification.RetryCount++;

                    // GetPendingNotificationsAsync only ever selects
                    // Status == Pending, so a failure that leaves this as
                    // Failed would never be retried — persist it back to
                    // Pending until the retry budget is actually exhausted.
                    var exhausted = notification.RetryCount >= _settings.MaxRetryAttempts;
                    notification.Status = exhausted
                        ? NotificationStatus.Failed.ToString()
                        : NotificationStatus.Pending.ToString();

                    if (exhausted)
                    {
                        _logger.LogWarning(
                            "Notification {NotificationId} exhausted retry attempts ({RetryCount}) and will not be retried again. Last error: {ErrorMessage}",
                            notificationId, notification.RetryCount, errorMessage);
                    }
                }
                else
                {
                    notification.Status = status.ToString();
                }

                await _context.SaveChangesAsync();
            }
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error updating notification status for ID {NotificationId}", notificationId);
        }
    }

    public async Task<IEnumerable<OverduePayment>> GetOverduePaymentsAsync()
    {
        try
        {
            var cutoffDate = DateOnly.FromDateTime(DateTime.UtcNow.AddDays(-5));
            var oneDayAgo = DateTime.UtcNow.AddDays(-1);

            var overduePayments = await _context.Payments
                .Include(p => p.Unit)
                .Include(p => p.Tenant)
                .Where(p => p.Status == PaymentStatus.Due &&
                           p.DueDate.HasValue &&
                           p.DueDate < cutoffDate)
                .Where(p => !_context.NotificationQueue
                    .Any(nq => nq.TenantId == p.TenantId &&
                              nq.UnitId == p.UnitId &&
                              nq.NotificationType == "Reminder" &&
                              nq.CreatedAt > oneDayAgo))
                .Select(p => new OverduePayment
                {
                    Id = p.Id,
                    UnitId = p.UnitId,
                    TenantId = p.TenantId,
                    Month = p.Month,
                    Year = p.Year,
                    AmountDue = p.AmountDue,
                    UnitName = p.Unit.UnitName ?? p.Unit.UnitNumber,
                    TenantName = p.Tenant.Name,
                    PhoneNumber = p.Tenant.PhoneNumber
                })
                .ToListAsync();

            return overduePayments;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error retrieving overdue payments");
            return Enumerable.Empty<OverduePayment>();
        }
    }

    public async Task CreateReminderNotificationAsync(int tenantId, int unitId, string message, NotificationType type)
    {
        try
        {
            var tenant = await _context.Tenants.FindAsync(tenantId);
            if (tenant == null) return;

            string? recipient = type switch
            {
                NotificationType.SMS => tenant.PhoneNumber,
                NotificationType.Email => tenant.Email,
                _ => tenant.PhoneNumber ?? tenant.Email
            };

            if (string.IsNullOrEmpty(recipient)) return;

            var notification = new NotificationQueue
            {
                TenantId = tenantId,
                UnitId = unitId,
                Message = message,
                NotificationType = type.ToString(),
                Recipient = recipient,
                Status = NotificationStatus.Pending.ToString(),
                CreatedAt = DateTime.UtcNow,
                UpdatedAt = DateTime.UtcNow
            };

            _context.NotificationQueue.Add(notification);
            await _context.SaveChangesAsync();
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error creating reminder notification");
        }
    }
}