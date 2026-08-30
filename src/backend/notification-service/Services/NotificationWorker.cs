using Microsoft.Extensions.Options;
using TenantlyNotificationService.Configuration;
using TenantlyNotificationService.Models;

namespace TenantlyNotificationService.Services;

public class NotificationWorker : BackgroundService
{
    private readonly IServiceProvider _serviceProvider;
    private readonly ISmsService _smsService;
    private readonly IEmailService _emailService;
    private readonly INotificationChannelListener _channelListener;
    private readonly NotificationSettings _settings;
    private readonly ILogger<NotificationWorker> _logger;

    public NotificationWorker(
        IServiceProvider serviceProvider,
        ISmsService smsService,
        IEmailService emailService,
        INotificationChannelListener channelListener,
        IOptions<NotificationSettings> settings,
        ILogger<NotificationWorker> logger)
    {
        _serviceProvider = serviceProvider;
        _smsService = smsService;
        _emailService = emailService;
        _channelListener = channelListener;
        _settings = settings.Value;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation("Notification Worker started");

        // ProcessingIntervalSeconds is now just the fallback poll cadence: the
        // worker normally wakes immediately via LISTEN/NOTIFY on inserts into
        // notification_queue, and only waits this long as a safety net if a
        // notification is missed (e.g. connection blip).
        var fallbackInterval = TimeSpan.FromSeconds(_settings.ProcessingIntervalSeconds);

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                await ProcessPendingNotifications();
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error in notification processing cycle");
            }

            if (stoppingToken.IsCancellationRequested)
            {
                break;
            }

            try
            {
                await _channelListener.WaitForNotificationAsync(fallbackInterval, stoppingToken);
            }
            catch (OperationCanceledException)
            {
                break;
            }
        }

        _logger.LogInformation("Notification Worker stopped");
    }

    private async Task ProcessPendingNotifications()
    {
        using var scope = _serviceProvider.CreateScope();
        var databaseService = scope.ServiceProvider.GetRequiredService<IDatabaseService>();

        var notifications = await databaseService.GetPendingNotificationsAsync();

        foreach (var notification in notifications)
        {
            try
            {
                bool success = false;

                if (Enum.TryParse<NotificationType>(notification.NotificationType, out var notificationType))
                {
                    switch (notificationType)
                    {
                        case NotificationType.SMS:
                            success = await _smsService.SendSmsAsync(notification.Recipient, notification.Message);
                            break;
                        case NotificationType.Email:
                            success = await _emailService.SendEmailAsync(notification.Recipient, "Rent Notification", notification.Message);
                            break;
                    }
                }

                var status = success ? NotificationStatus.Sent : NotificationStatus.Failed;
                var errorMessage = success ? null : "Delivery failed";

                await databaseService.UpdateNotificationStatusAsync(notification.Id, status, errorMessage);

                _logger.LogInformation("Processed notification {Id} with status {Status}", notification.Id, status);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error processing notification {Id}", notification.Id);
                await databaseService.UpdateNotificationStatusAsync(notification.Id, NotificationStatus.Failed, ex.Message);
            }
        }
    }
}