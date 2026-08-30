using Microsoft.Extensions.Options;
using TenantlyNotificationService.Configuration;
using TenantlyNotificationService.Models;

namespace TenantlyNotificationService.Services;

public class ReminderWorker : BackgroundService
{
    private readonly IServiceProvider _serviceProvider;
    private readonly NotificationSettings _settings;
    private readonly ILogger<ReminderWorker> _logger;

    public ReminderWorker(
        IServiceProvider serviceProvider,
        IOptions<NotificationSettings> settings,
        ILogger<ReminderWorker> logger)
    {
        _serviceProvider = serviceProvider;
        _settings = settings.Value;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation("Reminder Worker started");

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                await ProcessOverduePayments();

                // Run reminder check every hour
                await Task.Delay(TimeSpan.FromHours(1), stoppingToken);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error in reminder processing cycle");
                await Task.Delay(TimeSpan.FromMinutes(15), stoppingToken); // Wait shorter on error for reminders
            }
        }

        _logger.LogInformation("Reminder Worker stopped");
    }

    private async Task ProcessOverduePayments()
    {
        using var scope = _serviceProvider.CreateScope();
        var databaseService = scope.ServiceProvider.GetRequiredService<IDatabaseService>();

        var overduePayments = await databaseService.GetOverduePaymentsAsync();

        foreach (var payment in overduePayments)
        {
            try
            {
                var message = $"Dear {payment.TenantName}, your rent for {payment.UnitName} for {payment.Month}/{payment.Year} is overdue. Amount: ৳{payment.AmountDue}. Please pay as soon as possible.";

                // Create SMS reminder if phone number exists
                if (!string.IsNullOrEmpty(payment.PhoneNumber))
                {
                    await databaseService.CreateReminderNotificationAsync(
                        payment.TenantId,
                        payment.UnitId,
                        message,
                        NotificationType.SMS);
                }

                _logger.LogInformation("Created reminder for tenant {TenantId} unit {UnitId}", payment.TenantId, payment.UnitId);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error creating reminder for payment {PaymentId}", payment.Id);
            }
        }
    }
}