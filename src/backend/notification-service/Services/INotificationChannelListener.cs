namespace TenantlyNotificationService.Services;

public interface INotificationChannelListener
{
    /// <summary>
    /// Waits until either a Postgres NOTIFY arrives on the notification_queue
    /// channel or the timeout elapses. Returns true if woken by a notification,
    /// false on timeout. Never throws except on cancellation.
    /// </summary>
    Task<bool> WaitForNotificationAsync(TimeSpan timeout, CancellationToken cancellationToken);
}
