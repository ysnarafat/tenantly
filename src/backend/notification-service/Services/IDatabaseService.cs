using TenantlyNotificationService.Models;

namespace TenantlyNotificationService.Services;

public interface IDatabaseService
{
    Task<IEnumerable<NotificationQueue>> GetPendingNotificationsAsync();
    Task UpdateNotificationStatusAsync(int notificationId, NotificationStatus status, string? errorMessage = null);
    Task<IEnumerable<OverduePayment>> GetOverduePaymentsAsync();
    Task CreateReminderNotificationAsync(int tenantId, int unitId, string message, NotificationType type);
}