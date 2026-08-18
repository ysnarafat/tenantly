using Microsoft.Extensions.Logging.Abstractions;
using Microsoft.Extensions.Options;
using TenantlyNotificationService.Configuration;
using TenantlyNotificationService.Services;

namespace TenantlyNotificationService.Tests;

public class EmailServiceTests
{
    [Fact]
    public async Task SendEmailAsync_ReturnsFalse_WhenUsernameIsNotConfigured()
    {
        var settings = new NotificationSettings
        {
            Email = new EmailSettings { Username = string.Empty, Password = string.Empty },
        };
        var service = new EmailService(Options.Create(settings), NullLogger<EmailService>.Instance);

        var result = await service.SendEmailAsync("tenant@example.com", "Rent Notification", "body");

        Assert.False(result);
    }

    [Fact]
    public async Task SendEmailAsync_ReturnsFalse_WhenSmtpConnectionFails()
    {
        // Point at a host/port combination that refuses connections immediately,
        // so the send fails fast and we assert the failure is caught rather than thrown.
        var settings = new NotificationSettings
        {
            Email = new EmailSettings
            {
                SmtpHost = "127.0.0.1",
                SmtpPort = 1,
                Username = "someone@example.com",
                Password = "password",
                FromEmail = "noreply@tenantly.com",
                FromName = "Tenantly",
                EnableSsl = true,
            },
        };
        var service = new EmailService(Options.Create(settings), NullLogger<EmailService>.Instance);

        var result = await service.SendEmailAsync("tenant@example.com", "Rent Notification", "body");

        Assert.False(result);
    }
}
